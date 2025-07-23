package exporter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"fsd-crawler/pkg/config"
	"fsd-crawler/pkg/dependencies"
	"fsd-crawler/pkg/model"
)

func ExportJSON(structure *model.ProjectStructure, cfg *config.Config) error {
	outputDir := "./dist"
	if cfg != nil && cfg.OutputDir != "" {
		outputDir = cfg.OutputDir
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("не удалось создать директорию для вывода: %v", err)
	}

	outputPath := filepath.Join(outputDir, "fsd_structure.json")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("не удалось создать JSON файл: %v", err)
	}
	defer outputFile.Close()

	exportData := struct {
		Layers       []*model.FSDLayer       `json:"layers"`
		Dependencies []dependencies.Dependency `json:"dependencies"`
	}{
		Layers:       structure.Layers,
		Dependencies: []dependencies.Dependency{},
	}

	for _, dep := range structure.Dependencies {
		if d, ok := dep.(dependencies.Dependency); ok {
			exportData.Dependencies = append(exportData.Dependencies, d)
		}
	}

	encoder := json.NewEncoder(outputFile)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(exportData); err != nil {
		return fmt.Errorf("ошибка при кодировании в JSON: %v", err)
	}

	return nil
} 