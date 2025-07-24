package scripts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func BuildPlugins(pluginDir string) error {
	entries, err := os.ReadDir(pluginDir)
	if err != nil {
		return fmt.Errorf("failed to read plugin directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}

		inputPath := filepath.Join(pluginDir, entry.Name())
		outputName := strings.TrimSuffix(entry.Name(), ".go") + ".so"
		outputPath := filepath.Join(pluginDir, outputName)

		cmd := exec.Command("go", "build", "-buildmode=plugin", "-o", outputPath, inputPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		fmt.Printf("Building plugin: %s -> %s\n", inputPath, outputPath)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to build plugin %s: %w", entry.Name(), err)
		}
	}

	return nil
}
