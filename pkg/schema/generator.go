package schema

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

// Schema представляет JSON Schema
type Schema struct {
	Type        string                 `json:"type,omitempty"`
	Properties  map[string]interface{} `json:"properties,omitempty"`
	Required    []string               `json:"required,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Items       interface{}            `json:"items,omitempty"`
}

// Generator генерирует схемы для Go структур
type Generator struct {
	OutputDir string
}

// NewGenerator создает новый генератор схем
func NewGenerator(outputDir string) *Generator {
	return &Generator{
		OutputDir: outputDir,
	}
}

// GenerateSchema генерирует JSON Schema для типа
func (g *Generator) GenerateSchema(t reflect.Type, name string) (*Schema, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	schema := &Schema{
		Type:  getJSONType(t),
		Title: name,
	}

	if t.Kind() == reflect.Struct {
		schema.Properties = make(map[string]interface{})
		var required []string

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Пропускаем неэкспортируемые поля
			if !field.IsExported() {
				continue
			}

			fieldName := getJSONFieldName(field)
			fieldSchema := g.generateFieldSchema(field.Type)

			schema.Properties[fieldName] = fieldSchema

			// Добавляем в required, если поле не имеет тега omitempty
			jsonTag := field.Tag.Get("json")
			if !strings.Contains(jsonTag, "omitempty") && fieldName != "-" {
				required = append(required, fieldName)
			}
		}

		if len(required) > 0 {
			schema.Required = required
		}
	} else if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		itemType := t.Elem()
		schema.Items = g.generateFieldSchema(itemType)
	}

	return schema, nil
}

// generateFieldSchema генерирует схему для поля
func (g *Generator) generateFieldSchema(t reflect.Type) interface{} {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.Struct:
		nestedSchema := &Schema{
			Type:       "object",
			Properties: make(map[string]interface{}),
		}

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}

			fieldName := getJSONFieldName(field)
			nestedSchema.Properties[fieldName] = g.generateFieldSchema(field.Type)
		}

		return nestedSchema

	case reflect.Slice, reflect.Array:
		return &Schema{
			Type:  "array",
			Items: g.generateFieldSchema(t.Elem()),
		}

	default:
		return &Schema{
			Type: getJSONType(t),
		}
	}
}

// SaveSchema сохраняет схему в файл
func (g *Generator) SaveSchema(schema *Schema, filename string) error {
	// Создаем директорию если она не существует
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	filePath := filepath.Join(g.OutputDir, filename+".json")

	data, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal schema: %v", err)
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write schema file: %v", err)
	}

	fmt.Printf("Schema saved: %s\n", filePath)
	return nil
}

// GenerateForTypes генерирует схемы для списка типов
func (g *Generator) GenerateForTypes(types map[string]reflect.Type) error {
	for name, t := range types {
		schema, err := g.GenerateSchema(t, name)
		if err != nil {
			return fmt.Errorf("failed to generate schema for %s: %v", name, err)
		}

		err = g.SaveSchema(schema, strings.ToLower(name))
		if err != nil {
			return fmt.Errorf("failed to save schema for %s: %v", name, err)
		}
	}
	return nil
}

// getJSONType возвращает JSON тип для Go типа
func getJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	case reflect.Slice, reflect.Array:
		return "array"
	case reflect.Struct, reflect.Map:
		return "object"
	default:
		return "string"
	}
}

// getJSONFieldName получает имя поля из JSON тега или использует имя поля
func getJSONFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return strings.ToLower(field.Name)
	}

	// Парсим json тег
	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return "-"
	}
	if parts[0] != "" {
		return parts[0]
	}

	return strings.ToLower(field.Name)
}
