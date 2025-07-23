package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShort(t *testing.T) {
	t.Log("Short test passed")
}

func TestMain(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
		return
	}

	tempDir, err := os.MkdirTemp("", "fsd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	createTestFSDStructure(t, tempDir)

	outputDir := filepath.Join(tempDir, "dist")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	configContent := `
srcDir: "."
outputDir: "./dist"
outputFormats:
  - html
  - json
excludeDirs:
  - node_modules
  - .git
`
	configPath := filepath.Join(tempDir, "fsd-crawler.yml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}
	defer os.Chdir(currentDir)

	done := make(chan bool)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("main() panicked: %v", r)
			}
			done <- true
		}()
		main()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Log("main() did not complete in time, continuing with checks")
	}

	htmlPath := filepath.Join(outputDir, "fsd_structure.html")
	jsonPath := filepath.Join(outputDir, "fsd_structure.json")

	maxAttempts := 5
	for i := 0; i < maxAttempts; i++ {
		htmlExists := false
		jsonExists := false

		if _, err := os.Stat(htmlPath); err == nil {
			htmlExists = true
		}
		if _, err := os.Stat(jsonPath); err == nil {
			jsonExists = true
		}

		if htmlExists && jsonExists {
			return
		}

		t.Logf("Waiting for files to be created (attempt %d/%d)", i+1, maxAttempts)
		time.Sleep(500 * time.Millisecond)
	}

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Errorf("HTML file was not created in the specified output directory")
	}
	if _, err := os.Stat(jsonPath); os.IsNotExist(err) {
		t.Errorf("JSON file was not created in the specified output directory")
	}

	os.Remove(htmlPath)
	os.Remove(jsonPath)

	configContent = `
srcDir: "."
outputDir: "./dist"
outputFormats:
  - html
excludeDirs:
  - node_modules
  - .git
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("main() panicked: %v", r)
			}
			done <- true
		}()
		main()
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Log("main() did not complete in time, continuing with checks")
	}

	if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
		t.Errorf("HTML file was not created with custom config")
	}
	if _, err := os.Stat(jsonPath); err == nil {
		t.Errorf("JSON file was created when it should not have been")
	}
}

func createTestFSDStructure(t *testing.T, rootDir string) {
	dirs := []string{
		"app/routes",
		"app/config",
		"entities/user/api",
		"entities/user/model",
		"features/auth/ui",
		"features/auth/model",
		"pages/home/ui",
		"shared/ui",
		"shared/api",
	}

	files := map[string]string{
		"app/routes/routes.ts":        "",
		"app/config/config.ts":        "",
		"entities/user/api/userApi.ts": "",
		"entities/user/model/user.ts":  "",
		"features/auth/ui/login.tsx":   "",
		"features/auth/model/auth.ts":  "",
		"pages/home/ui/HomePage.tsx":   "",
		"shared/ui/Button.tsx":         "",
		"shared/api/api.ts":            "",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(rootDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}
	}

	for path, content := range files {
		filePath := filepath.Join(rootDir, path)
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}
} 