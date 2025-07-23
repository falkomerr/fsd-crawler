package config

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "fsd-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем тестовый конфиг
	configContent := `
srcDir: ./src
outputDir: ./output
outputFormats:
  - html
  - json
  - md
excludeDirs:
  - node_modules
  - .git
  - dist
customLayers:
  - app
  - pages
  - widgets
htmlTemplatePath: ./custom-template.html
`
	configPath := filepath.Join(tempDir, "fsd-crawler.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Загружаем конфиг
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Проверяем значения
	expectedConfig := &Config{
		SrcDir:          "./src",
		OutputDir:        "./output",
		OutputFormats:    []string{"html", "json", "md"},
		ExcludeDirs:      []string{"node_modules", ".git", "dist"},
		CustomLayers:     []string{"app", "pages", "widgets"},
		HTMLTemplatePath: "./custom-template.html",
	}

	if config.SrcDir != expectedConfig.SrcDir {
		t.Errorf("SrcDir = %s; want %s", config.SrcDir, expectedConfig.SrcDir)
	}

	if config.OutputDir != expectedConfig.OutputDir {
		t.Errorf("OutputDir = %s; want %s", config.OutputDir, expectedConfig.OutputDir)
	}

	if !reflect.DeepEqual(config.OutputFormats, expectedConfig.OutputFormats) {
		t.Errorf("OutputFormats = %v; want %v", config.OutputFormats, expectedConfig.OutputFormats)
	}

	if !reflect.DeepEqual(config.CustomLayers, expectedConfig.CustomLayers) {
		t.Errorf("CustomLayers = %v; want %v", config.CustomLayers, expectedConfig.CustomLayers)
	}

	if config.HTMLTemplatePath != expectedConfig.HTMLTemplatePath {
		t.Errorf("HTMLTemplatePath = %s; want %s", config.HTMLTemplatePath, expectedConfig.HTMLTemplatePath)
	}
}

func TestFindAndLoadConfig(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "fsd-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем вложенную структуру директорий
	nestedDir := filepath.Join(tempDir, "project", "src")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Создаем тестовый конфиг в корне проекта
	configContent := `
srcDir: ./src
outputDir: ./output
outputFormats:
  - html
  - json
excludeDirs:
  - node_modules
  - .git
`
	configPath := filepath.Join(tempDir, "project", "fsd-crawler.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Сохраняем текущую директорию
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Переходим во вложенную директорию
	if err := os.Chdir(nestedDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(currentDir)

	// Пытаемся найти и загрузить конфиг
	config, err := FindAndLoadConfig()
	if err != nil {
		t.Fatalf("FindAndLoadConfig failed: %v", err)
	}

	// Проверяем, что конфиг был найден и загружен
	if config.OutputDir != "./output" {
		t.Errorf("OutputDir = %s; want ./output", config.OutputDir)
	}

	if !reflect.DeepEqual(config.OutputFormats, []string{"html", "json"}) {
		t.Errorf("OutputFormats = %v; want [html json]", config.OutputFormats)
	}

	// Проверяем, что если конфиг не найден, возвращаются значения по умолчанию
	// Для этого удаляем конфиг и пробуем снова
	if err := os.Remove(configPath); err != nil {
		t.Fatalf("Failed to remove config file: %v", err)
	}

	config, err = FindAndLoadConfig()
	if err != nil {
		t.Fatalf("FindAndLoadConfig failed: %v", err)
	}

	// Проверяем, что вернулись значения по умолчанию
	if !reflect.DeepEqual(config, &DefaultConfig) {
		t.Errorf("Default config not returned when no config file found")
	}
} 