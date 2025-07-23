package exporter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fsd-crawler/pkg/config"
	"fsd-crawler/pkg/model"
)

func createTestStructure() *model.ProjectStructure {
	// Создаем тестовую структуру проекта
	return &model.ProjectStructure{
		Layers: []*model.FSDLayer{
			{
				Name: "app",
				Slices: []*model.FSDSlice{
					{
						Name: "",
						Segments: []*model.FSDSegment{
							{
								Name:  "routes",
								Files: []string{"routes.ts"},
							},
						},
					},
				},
			},
			{
				Name: "entities",
				Slices: []*model.FSDSlice{
					{
						Name: "user",
						Segments: []*model.FSDSegment{
							{
								Name:  "api",
								Files: []string{"userApi.ts"},
							},
							{
								Name:  "model",
								Files: []string{"user.ts"},
							},
						},
					},
				},
			},
		},
	}
}

func TestGenerateHTML(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "fsd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем поддиректорию для вывода
	outputDir := filepath.Join(tempDir, "dist")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Сохраняем текущую директорию
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Переходим во временную директорию
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(currentDir)

	// Создаем тестовую структуру
	structure := createTestStructure()

	// Создаем конфигурацию
	cfg := &config.Config{
		OutputDir: outputDir,
	}

	// Генерируем HTML с конфигурацией
	if err := GenerateHTML(structure, cfg); err != nil {
		t.Fatalf("GenerateHTML failed: %v", err)
	}

	// Проверяем, что файл создан в указанной директории
	htmlPath := filepath.Join(outputDir, "fsd_structure.html")
	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Errorf("HTML file was not created in the specified output directory")
	}

	// Читаем содержимое файла
	content, err := os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	// Проверяем, что файл содержит ожидаемые данные
	htmlContent := string(content)
	expectedStrings := []string{
		"Feature-Sliced Design Structure Analyzer",
		"Слой: app",
		"Слой: entities",
		"Слайс: user",
		"Сегмент: routes",
		"Сегмент: api",
		"Сегмент: model",
		"routes.ts",
		"userApi.ts",
		"user.ts",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(htmlContent, expected) {
			t.Errorf("HTML file does not contain expected string: %s", expected)
		}
	}

	// Тест с пользовательским шаблоном
	customTemplate := `
<!DOCTYPE html>
<html>
<head>
    <title>Custom Template</title>
</head>
<body>
    <h1>Custom FSD Report</h1>
    <ul>
    {{range .Layers}}
        <li>{{.Name}}</li>
    {{end}}
    </ul>
</body>
</html>
`
	customTemplatePath := filepath.Join(tempDir, "custom-template.html")
	if err := os.WriteFile(customTemplatePath, []byte(customTemplate), 0644); err != nil {
		t.Fatalf("Failed to write custom template: %v", err)
	}

	// Обновляем конфигурацию
	cfg.HTMLTemplatePath = customTemplatePath

	// Генерируем HTML с пользовательским шаблоном
	if err := GenerateHTML(structure, cfg); err != nil {
		t.Fatalf("GenerateHTML with custom template failed: %v", err)
	}

	// Читаем содержимое файла
	content, err = os.ReadFile(htmlPath)
	if err != nil {
		t.Fatalf("Failed to read HTML file: %v", err)
	}

	// Проверяем, что файл содержит данные из пользовательского шаблона
	htmlContent = string(content)
	if !strings.Contains(htmlContent, "Custom FSD Report") {
		t.Errorf("HTML file does not contain custom template content")
	}
}

func TestExportJSON(t *testing.T) {
	// Создаем временную директорию для тестов
	tempDir, err := os.MkdirTemp("", "fsd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем поддиректорию для вывода
	outputDir := filepath.Join(tempDir, "dist")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	// Сохраняем текущую директорию
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Переходим во временную директорию
	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(currentDir)

	// Создаем тестовую структуру
	structure := createTestStructure()

	// Создаем конфигурацию
	cfg := &config.Config{
		OutputDir: outputDir,
	}

	// Экспортируем в JSON с конфигурацией
	if err := ExportJSON(structure, cfg); err != nil {
		t.Fatalf("ExportJSON failed: %v", err)
	}

	// Проверяем, что файл создан в указанной директории
	jsonPath := filepath.Join(outputDir, "fsd_structure.json")
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Errorf("JSON file was not created in the specified output directory")
	}

	// Читаем содержимое файла
	content, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("Failed to read JSON file: %v", err)
	}

	// Декодируем JSON
	var decodedStructure model.ProjectStructure
	if err := json.Unmarshal(content, &decodedStructure); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	// Проверяем, что данные соответствуют исходной структуре
	if len(decodedStructure.Layers) != len(structure.Layers) {
		t.Errorf("JSON Layers length = %d; want %d", len(decodedStructure.Layers), len(structure.Layers))
	}

	// Проверяем первый слой
	if len(decodedStructure.Layers) > 0 {
		if decodedStructure.Layers[0].Name != "app" {
			t.Errorf("JSON first layer name = %s; want app", decodedStructure.Layers[0].Name)
		}
	}

	// Проверяем второй слой
	if len(decodedStructure.Layers) > 1 {
		if decodedStructure.Layers[1].Name != "entities" {
			t.Errorf("JSON second layer name = %s; want entities", decodedStructure.Layers[1].Name)
		}

		// Проверяем слайс user
		if len(decodedStructure.Layers[1].Slices) > 0 {
			if decodedStructure.Layers[1].Slices[0].Name != "user" {
				t.Errorf("JSON user slice name = %s; want user", decodedStructure.Layers[1].Slices[0].Name)
			}
		}
	}
} 