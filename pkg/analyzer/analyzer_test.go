package analyzer

import (
	"os"
	"path/filepath"
	"testing"

	"fsd-crawler/pkg/config"
)

func TestIsExcluded(t *testing.T) {
	testCases := []struct {
		name        string
		excludeDirs []string
		expected    bool
	}{
		{"node_modules", []string{"node_modules", ".git", "dist"}, true},
		{".git", []string{"node_modules", ".git", "dist"}, true},
		{".gitignore", []string{"node_modules", ".git", "dist"}, true}, // Начинается с точки
		{"src", []string{"node_modules", ".git", "dist"}, false},
		{"app", []string{"node_modules", ".git", "dist"}, false},
	}

	for _, tc := range testCases {
		result := isExcluded(tc.name, tc.excludeDirs)
		if result != tc.expected {
			t.Errorf("isExcluded(%s, %v) = %v; want %v", 
				tc.name, tc.excludeDirs, result, tc.expected)
		}
	}
}

func TestContains(t *testing.T) {
	testCases := []struct {
		slice    []string
		item     string
		expected bool
	}{
		{[]string{"a", "b", "c"}, "a", true},
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "c", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "a", false},
	}

	for _, tc := range testCases {
		result := contains(tc.slice, tc.item)
		if result != tc.expected {
			t.Errorf("contains(%v, %s) = %v; want %v", 
				tc.slice, tc.item, result, tc.expected)
		}
	}
}

func TestFindFiles(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fsd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Создаем файлы и директории для тестирования
	files := []string{
		"file1.ts",
		"file2.tsx",
		"subdir/file3.js",
		"subdir/file4.jsx",
		".hidden/file5.ts",
		"node_modules/file6.ts",
	}

	for _, file := range files {
		filePath := filepath.Join(tempDir, file)
		dir := filepath.Dir(filePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(filePath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", filePath, err)
		}
	}

	cfg := &config.Config{
		ExcludeDirs: []string{"node_modules", ".git", ".hidden"},
	}

	foundFiles := findFiles(tempDir, cfg)

	// Проверяем, что найдены правильные файлы
	expectedFiles := []string{
		"file1.ts",
		"file2.tsx",
		"subdir/file3.js",
		"subdir/file4.jsx",
	}

	if len(foundFiles) != len(expectedFiles) {
		t.Errorf("findFiles found %d files; want %d", len(foundFiles), len(expectedFiles))
	}

	// Проверяем, что все ожидаемые файлы найдены
	for _, expected := range expectedFiles {
		found := false
		for _, actual := range foundFiles {
			if filepath.ToSlash(actual) == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected file %s not found", expected)
		}
	}

	// Проверяем, что исключенные файлы не найдены
	excludedFiles := []string{
		".hidden/file5.ts",
		"node_modules/file6.ts",
	}

	for _, excluded := range excludedFiles {
		for _, actual := range foundFiles {
			if filepath.ToSlash(actual) == excluded {
				t.Errorf("Excluded file %s was found", excluded)
			}
		}
	}
} 