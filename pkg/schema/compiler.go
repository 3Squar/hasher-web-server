package schema

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Compiler компилирует FlatBuffer схемы в Go код
type Compiler struct {
	SchemaDir string
	OutputDir string
}

// NewCompiler создает новый компилятор FlatBuffer схем
func NewCompiler(schemaDir, outputDir string) *Compiler {
	return &Compiler{
		SchemaDir: schemaDir,
		OutputDir: outputDir,
	}
}

// CompileSchemas компилирует все .fbs файлы в Go код
func (c *Compiler) CompileSchemas() error {
	// Создаем выходную директорию если она не существует
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Находим все .fbs файлы
	fbsFiles, err := c.findFBSFiles()
	if err != nil {
		return fmt.Errorf("failed to find .fbs files: %v", err)
	}

	if len(fbsFiles) == 0 {
		fmt.Println("No .fbs files found to compile")
		return nil
	}

	fmt.Printf("Found %d .fbs files to compile...\n", len(fbsFiles))

	// Компилируем каждый файл
	for _, fbsFile := range fbsFiles {
		err := c.compileFBSFile(fbsFile)
		if err != nil {
			return fmt.Errorf("failed to compile %s: %v", fbsFile, err)
		}
		fmt.Printf("Compiled: %s\n", fbsFile)
	}

	fmt.Println("All FlatBuffer schemas compiled successfully!")
	return nil
}

// findFBSFiles находит все .fbs файлы в директории схем
func (c *Compiler) findFBSFiles() ([]string, error) {
	var fbsFiles []string

	err := filepath.Walk(c.SchemaDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".fbs") {
			fbsFiles = append(fbsFiles, path)
		}

		return nil
	})

	return fbsFiles, err
}

// compileFBSFile компилирует один .fbs файл в Go код
func (c *Compiler) compileFBSFile(fbsFile string) error {
	// Команда: flatc --go -o output_dir schema.fbs
	cmd := exec.Command("flatc", "--go", "-o", c.OutputDir, fbsFile)

	// Устанавливаем рабочую директорию
	cmd.Dir = "."

	// Выполняем команду
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("flatc command failed: %v\nOutput: %s", err, string(output))
	}

	return nil
}

// CompileSchemasWithPackage компилирует схемы с указанием пакета Go
func (c *Compiler) CompileSchemasWithPackage(packageName string) error {
	// Создаем выходную директорию если она не существует
	if err := os.MkdirAll(c.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Находим все .fbs файлы
	fbsFiles, err := c.findFBSFiles()
	if err != nil {
		return fmt.Errorf("failed to find .fbs files: %v", err)
	}

	if len(fbsFiles) == 0 {
		fmt.Println("No .fbs files found to compile")
		return nil
	}

	fmt.Printf("Found %d .fbs files to compile with package '%s'...\n", len(fbsFiles), packageName)

	// Компилируем все файлы одной командой для лучшей обработки зависимостей
	args := []string{"--go", "-o", c.OutputDir}
	if packageName != "" {
		args = append(args, "--go-namespace", packageName)
	}
	args = append(args, fbsFiles...)

	cmd := exec.Command("flatc", args...)
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("flatc command failed: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("Output from flatc:\n%s\n", string(output))
	fmt.Println("All FlatBuffer schemas compiled successfully!")
	return nil
}

// CheckFlatc проверяет доступность утилиты flatc
func CheckFlatc() error {
	cmd := exec.Command("flatc", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("flatc not found or not working: %v\nPlease install FlatBuffers: brew install flatbuffers", err)
	}

	fmt.Printf("FlatBuffer compiler version: %s", string(output))
	return nil
}
