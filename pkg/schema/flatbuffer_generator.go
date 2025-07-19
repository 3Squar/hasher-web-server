package schema

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

// FlatBufferGenerator генерирует FlatBuffer схемы (.fbs файлы)
type FlatBufferGenerator struct {
	OutputDir string
	Namespace string
}

// NewFlatBufferGenerator создает новый генератор FlatBuffer схем
func NewFlatBufferGenerator(outputDir, namespace string) *FlatBufferGenerator {
	return &FlatBufferGenerator{
		OutputDir: outputDir,
		Namespace: namespace,
	}
}

// GenerateFlatBufferSchema генерирует FlatBuffer схему для типа
func (g *FlatBufferGenerator) GenerateFlatBufferSchema(t reflect.Type, name string) (string, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var schema strings.Builder

	// Добавляем namespace
	if g.Namespace != "" {
		schema.WriteString(fmt.Sprintf("namespace %s;\n\n", g.Namespace))
	}

	// Генерируем определения вложенных типов, исключая основной тип
	dependencies := g.collectDependencies(t)
	processed := make(map[string]bool)

	for _, dep := range dependencies {
		if dep.Name != name && !processed[dep.Name] {
			depSchema := g.generateTypeDefinition(dep.Type, dep.Name)
			schema.WriteString(depSchema)
			schema.WriteString("\n")
			processed[dep.Name] = true
		}
	}

	// Генерируем основной тип
	mainSchema := g.generateTypeDefinition(t, name)
	schema.WriteString(mainSchema)

	return schema.String(), nil
}

// Dependency представляет зависимость типа
type Dependency struct {
	Name string
	Type reflect.Type
}

// collectDependencies собирает все зависимые типы
func (g *FlatBufferGenerator) collectDependencies(t reflect.Type) []Dependency {
	var deps []Dependency
	visited := make(map[reflect.Type]bool)

	g.collectDependenciesRecursive(t, visited, &deps)

	return deps
}

func (g *FlatBufferGenerator) collectDependenciesRecursive(t reflect.Type, visited map[reflect.Type]bool, deps *[]Dependency) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if visited[t] || t.Kind() != reflect.Struct {
		return
	}

	// Пропускаем встроенные типы
	if isBuiltInType(t) {
		return
	}

	visited[t] = true

	// Добавляем зависимость только если это не основной тип
	typeName := t.Name()
	if typeName != "" {
		*deps = append(*deps, Dependency{Name: typeName, Type: t})
	}

	// Рекурсивно обрабатываем поля
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Slice || fieldType.Kind() == reflect.Array {
			fieldType = fieldType.Elem()
		}

		g.collectDependenciesRecursive(fieldType, visited, deps)
	}
}

// generateTypeDefinition генерирует определение типа для FlatBuffer
func (g *FlatBufferGenerator) generateTypeDefinition(t reflect.Type, name string) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return ""
	}

	// Проверяем есть ли экспортируемые поля
	hasFields := false
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.IsExported() {
			fieldName := getFlatBufferFieldName(field)
			if fieldName != "-" {
				hasFields = true
				break
			}
		}
	}

	// Пропускаем пустые структуры (FlatBuffer не поддерживает структуры без полей)
	if !hasFields {
		fmt.Printf("Skipping empty struct: %s\n", name)
		return ""
	}

	var schema strings.Builder

	// Определяем тип структуры (table или struct)
	// В FlatBuffer table используется для сложных типов, struct для простых
	isSimpleStruct := g.isSimpleStruct(t)

	if isSimpleStruct {
		schema.WriteString(fmt.Sprintf("struct %s {\n", name))
	} else {
		schema.WriteString(fmt.Sprintf("table %s {\n", name))
	}

	// Добавляем поля
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		// Обрабатываем встроенные поля (anonymous fields)
		if field.Anonymous {
			// Для встроенных структур добавляем их поля напрямую
			embeddedType := field.Type
			if embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}

			if embeddedType.Kind() == reflect.Struct {
				for j := 0; j < embeddedType.NumField(); j++ {
					embeddedField := embeddedType.Field(j)
					if !embeddedField.IsExported() {
						continue
					}

					embeddedFieldName := getFlatBufferFieldName(embeddedField)
					if embeddedFieldName == "-" {
						continue
					}

					embeddedFieldType := g.getFlatBufferType(embeddedField.Type)
					schema.WriteString(fmt.Sprintf("  %s: %s;\n", embeddedFieldName, embeddedFieldType))
				}
			}
			continue
		}

		fieldName := getFlatBufferFieldName(field)
		if fieldName == "-" {
			continue
		}

		fieldType := g.getFlatBufferType(field.Type)
		schema.WriteString(fmt.Sprintf("  %s: %s;\n", fieldName, fieldType))
	}

	schema.WriteString("}\n")

	return schema.String()
}

// isSimpleStruct определяет является ли структура простой (только примитивные типы)
func (g *FlatBufferGenerator) isSimpleStruct(t reflect.Type) bool {
	// Если есть поля с встроенными структурами без полей, считаем простой
	if t.NumField() == 0 {
		return true
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		// Если есть сложные типы, это не простая структура
		switch fieldType.Kind() {
		case reflect.Struct:
			// Встроенные структуры (анонимные поля) обрабатываем особо
			if field.Anonymous {
				// Для встроенных структур проверяем их поля
				if !g.isSimpleStruct(fieldType) {
					return false
				}
			} else if !isBuiltInType(fieldType) {
				return false
			}
		case reflect.Slice, reflect.Array, reflect.Map:
			return false
		case reflect.String:
			// В FlatBuffer string поля делают структуру table, а не struct
			return false
		}
	}
	return true
}

// getFlatBufferType возвращает FlatBuffer тип для Go типа
func (g *FlatBufferGenerator) getFlatBufferType(t reflect.Type) string {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int8:
		return "int8"
	case reflect.Int16:
		return "int16"
	case reflect.Int32, reflect.Int:
		return "int32"
	case reflect.Int64:
		return "int64"
	case reflect.Uint8:
		return "uint8"
	case reflect.Uint16:
		return "uint16"
	case reflect.Uint32, reflect.Uint:
		return "uint32"
	case reflect.Uint64:
		return "uint64"
	case reflect.Float32:
		return "float32"
	case reflect.Float64:
		return "float64"
	case reflect.Bool:
		return "bool"
	case reflect.Slice, reflect.Array:
		elemType := g.getFlatBufferType(t.Elem())
		return fmt.Sprintf("[%s]", elemType)
	case reflect.Struct:
		if isBuiltInType(t) {
			// Для time.Time используем int64 (Unix timestamp)
			if t == reflect.TypeOf(time.Time{}) {
				return "int64"
			}
			return "string" // Для других встроенных типов
		}
		return t.Name()
	default:
		return "string"
	}
}

// getFlatBufferFieldName получает имя поля из тегов или использует имя поля
func getFlatBufferFieldName(field reflect.StructField) string {
	// Сначала проверяем flatbuffer тег
	if fbTag := field.Tag.Get("flatbuffer"); fbTag != "" {
		parts := strings.Split(fbTag, ",")
		if parts[0] == "-" {
			return "-"
		}
		if parts[0] != "" {
			return parts[0]
		}
	}

	// Затем проверяем json тег
	if jsonTag := field.Tag.Get("json"); jsonTag != "" {
		parts := strings.Split(jsonTag, ",")
		if parts[0] == "-" {
			return "-"
		}
		if parts[0] != "" {
			return parts[0]
		}
	}

	// Используем имя поля в snake_case
	return toSnakeCase(field.Name)
}

// toSnakeCase конвертирует CamelCase в snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// isBuiltInType проверяет является ли тип встроенным
func isBuiltInType(t reflect.Type) bool {
	switch t {
	case reflect.TypeOf(time.Time{}):
		return true
	default:
		return t.PkgPath() == ""
	}
}

// SaveFlatBufferSchema сохраняет FlatBuffer схему в файл
func (g *FlatBufferGenerator) SaveFlatBufferSchema(schema, filename string) error {
	// Создаем директорию если она не существует
	if err := os.MkdirAll(g.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	filePath := filepath.Join(g.OutputDir, filename+".fbs")

	err := os.WriteFile(filePath, []byte(schema), 0644)
	if err != nil {
		return fmt.Errorf("failed to write schema file: %v", err)
	}

	fmt.Printf("FlatBuffer schema saved: %s\n", filePath)
	return nil
}

// GenerateForTypes генерирует FlatBuffer схемы для списка типов
func (g *FlatBufferGenerator) GenerateForTypes(types map[string]reflect.Type) error {
	for name, t := range types {
		schema, err := g.GenerateFlatBufferSchema(t, name)
		if err != nil {
			return fmt.Errorf("failed to generate FlatBuffer schema for %s: %v", name, err)
		}

		// Пропускаем пустые схемы
		if strings.TrimSpace(schema) == "" || strings.Contains(schema, fmt.Sprintf("namespace %s;\n\n", g.Namespace)) && len(strings.TrimSpace(schema)) == len(fmt.Sprintf("namespace %s;", g.Namespace))+2 {
			fmt.Printf("Skipping empty schema for: %s\n", name)
			continue
		}

		err = g.SaveFlatBufferSchema(schema, strings.ToLower(name))
		if err != nil {
			return fmt.Errorf("failed to save FlatBuffer schema for %s: %v", name, err)
		}
	}
	return nil
}
