package dependencies

import (
	"os"
	"path/filepath"
	"testing"

	"fsd-crawler/pkg/config"
	"fsd-crawler/pkg/model"
)

func TestDetermineDependencyType(t *testing.T) {
	structure := &model.ProjectStructure{}
	analyzer := NewDependencyAnalyzer(structure, "", nil)

	testCases := []struct {
		fromLayer string
		toLayer   string
		expected  DependencyType
	}{
		{"app", "features", DependencyNormal},
		{"app", "entities", DependencyNormal},
		{"features", "entities", DependencyNormal},
		{"widgets", "entities", DependencyNormal},
		{"pages", "widgets", DependencyNormal},
		
		{"entities", "entities", DependencySameLayer},
		{"features", "features", DependencySameLayer},
		
		{"entities", "features", DependencyCyclical},
		{"entities", "widgets", DependencyCyclical},
		{"entities", "pages", DependencyCyclical},
		{"features", "pages", DependencyCyclical},
		{"widgets", "pages", DependencyCyclical},
	}

	for _, tc := range testCases {
		result := analyzer.determineDependencyType(tc.fromLayer, tc.toLayer)
		if result != tc.expected {
			t.Errorf("determineDependencyType(%s, %s) = %s; want %s", 
				tc.fromLayer, tc.toLayer, result, tc.expected)
		}
	}
}

func TestExtractLayerAndSlice(t *testing.T) {
	structure := &model.ProjectStructure{}
	analyzer := NewDependencyAnalyzer(structure, "", nil)

	testCases := []struct {
		importPath   string
		expectedLayer string
		expectedSlice string
	}{
		{"entities/user/api", "entities", "user/api"},
		{"features/auth/ui", "features", "auth/ui"},
		{"app/routes", "app", "routes"},
		{"shared/ui", "shared", "ui"},
		
		{"../entities/user/api", "entities", "user/api"},
		{"./features/auth/ui", "features", "auth/ui"},
		{"app/routes", "app", "routes"},
		
		{"utils/helpers", "", ""},
		{"@testing/library", "", ""},
	}

	for _, tc := range testCases {
		layer, slice := analyzer.extractLayerAndSlice(tc.importPath)
		if layer != tc.expectedLayer || slice != tc.expectedSlice {
			t.Errorf("extractLayerAndSlice(%s) = (%s, %s); want (%s, %s)", 
				tc.importPath, layer, slice, tc.expectedLayer, tc.expectedSlice)
		}
	}
}

func TestResolveAliasPath(t *testing.T) {
	structure := &model.ProjectStructure{}
	cfg := &config.Config{
		Aliases: map[string]string{
			"@": "src",
			"~": "app",
		},
	}
	analyzer := NewDependencyAnalyzer(structure, "", cfg)

	testCases := []struct {
		importPath   string
		expectedPath string
	}{
		{"@/features/auth", "src/features/auth"},
		{"~/components", "app/components"},
		{"features/auth", "features/auth"},
	}

	for _, tc := range testCases {
		result := analyzer.resolveAliasPath(tc.importPath)
		if result != tc.expectedPath {
			t.Errorf("resolveAliasPath(%s) = %s; want %s", 
				tc.importPath, result, tc.expectedPath)
		}
	}
}

func TestAnalyzeFileImports(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "fsd-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.ts")
	content := `
import { userModel } from 'entities/user/model';
import { Button } from 'shared/ui/Button';
import { HomePage } from 'pages/home/ui';
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	structure := &model.ProjectStructure{}
	analyzer := NewDependencyAnalyzer(structure, "", nil)

	analyzer.analyzeFileImports(testFile, "features", "login")

	expectedDependencies := map[string]string{
		"shared": "normal",
		"entities": "normal",
		"pages": "cyclical",
	}

	foundDeps := make(map[string]bool)
	for _, dep := range analyzer.dependencies {
		foundDeps[dep.ToLayer] = true
		
		expectedType := DependencyType(expectedDependencies[dep.ToLayer])
		if dep.Type != expectedType {
			t.Errorf("Dependency to %s has type %s; want %s", 
				dep.ToLayer, dep.Type, expectedType)
		}
		
		if dep.FromLayer != "features" || dep.FromSlice != "login" {
			t.Errorf("Dependency from %s/%s; want features/login", 
				dep.FromLayer, dep.FromSlice)
		}
	}

	for layer := range expectedDependencies {
		if !foundDeps[layer] {
			t.Errorf("Expected dependency to %s not found", layer)
		}
	}
}

func TestGetProblematicDependencies(t *testing.T) {
	structure := &model.ProjectStructure{}
	analyzer := NewDependencyAnalyzer(structure, "", nil)

	analyzer.dependencies = []Dependency{
		{FromLayer: "entities", FromSlice: "user", ToLayer: "features", ToSlice: "auth", Type: DependencyNormal},
		{FromLayer: "features", FromSlice: "auth", ToLayer: "entities", ToSlice: "user", Type: DependencyCyclical},
		{FromLayer: "features", FromSlice: "auth", ToLayer: "features", ToSlice: "profile", Type: DependencySameLayer},
		{FromLayer: "pages", FromSlice: "home", ToLayer: "widgets", ToSlice: "header", Type: DependencyNormal},
	}

	problematic := analyzer.GetProblematicDependencies()

	if len(problematic) != 1 {
		t.Errorf("Found %d problematic dependencies; want 1", len(problematic))
	}

	for _, dep := range problematic {
		if dep.Type != DependencyCyclical {
			t.Errorf("Dependency type %s is not problematic", dep.Type)
		}
	}
} 