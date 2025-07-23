package dependencies

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"fsd-crawler/pkg/config"
	"fsd-crawler/pkg/model"
)

type DependencyType string

const (
	DependencyNormal   DependencyType = "normal"
	DependencySameLayer DependencyType = "same"
	DependencyCyclical DependencyType = "cyclical"
	DependencyTest     DependencyType = "test"
)

type Dependency struct {
	FromLayer string
	FromSlice string
	ToLayer   string
	ToSlice   string
	Type      DependencyType
}

type DependencyAnalyzer struct {
	structure    *model.ProjectStructure
	rootDir      string
	layerIndices map[string]int
	dependencies []Dependency
	DetermineDepType func(fromLayer, fromSlice, toLayer, toSlice string) DependencyType
	config       *config.Config
}

func NewDependencyAnalyzer(structure *model.ProjectStructure, rootDir string, cfg *config.Config) *DependencyAnalyzer {
	layerIndices := make(map[string]int)
	for i, layerName := range model.KnownLayers {
		layerIndices[layerName] = i
	}

	da := &DependencyAnalyzer{
		structure:    structure,
		rootDir:      rootDir,
		layerIndices: layerIndices,
		dependencies: []Dependency{},
		config:       cfg,
	}
	
	da.DetermineDepType = da.determineDependencyTypeWithSlices
	
	return da
}

func (da *DependencyAnalyzer) AnalyzeDependencies() []Dependency {
	da.dependencies = []Dependency{}

	for _, layer := range da.structure.Layers {
		for _, slice := range layer.Slices {
			sliceName := slice.Name
			if sliceName == "" {
				sliceName = layer.Name
			}
			da.analyzeSliceImports(layer.Name, slice)
		}
	}

	return da.dependencies
}

func (da *DependencyAnalyzer) analyzeSliceImports(layerName string, slice *model.FSDSlice) {
	sliceName := slice.Name
	if sliceName == "" {
		sliceName = layerName
	}

	for _, segment := range slice.Segments {
		segmentPath := filepath.Join(da.rootDir, layerName)
		if sliceName != layerName {
			segmentPath = filepath.Join(segmentPath, sliceName)
		}
		segmentPath = filepath.Join(segmentPath, segment.Name)

		for _, file := range segment.Files {
			filePath := filepath.Join(segmentPath, file)
			da.analyzeFileImports(filePath, layerName, sliceName)
		}
	}
}

func (da *DependencyAnalyzer) analyzeFileImports(filePath, fromLayer, fromSlice string) {
	supportedExtensions := []string{".js", ".jsx", ".ts", ".tsx"}

	ext := filepath.Ext(filePath)
	supported := false
	for _, supportedExt := range supportedExtensions {
		if ext == supportedExt {
			supported = true
			break
		}
	}
	if !supported {
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	importPatterns := []*regexp.Regexp{
		regexp.MustCompile(`import\s+.*\s+from\s+['"]([^'"]+)['"]\s*;?`),
		regexp.MustCompile(`import\s+['"]([^'"]+)['"]\s*;?`),
		regexp.MustCompile(`require\s*\(\s*['"]([^'"]+)['"]\s*\)`),
	}

	lineNum := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		for _, pattern := range importPatterns {
			matches := pattern.FindStringSubmatch(line)
			if len(matches) > 1 {
				importPath := matches[1]
				
				resolvedPath := da.resolveAliasPath(importPath)
				
				toLayer, toSlice := da.extractLayerAndSlice(resolvedPath)
				
				if toLayer != "" {
					depType := da.DetermineDepType(fromLayer, fromSlice, toLayer, toSlice)
					
					dependency := Dependency{
						FromLayer: fromLayer,
						FromSlice: fromSlice,
						ToLayer:   toLayer,
						ToSlice:   toSlice,
						Type:      depType,
					}
					da.dependencies = append(da.dependencies, dependency)
				}
			}
		}
	}
}

func (da *DependencyAnalyzer) resolveAliasPath(importPath string) string {
	if da.config == nil || len(da.config.Aliases) == 0 {
		return importPath
	}

	for alias, target := range da.config.Aliases {
		if strings.HasPrefix(importPath, alias) {
			resolvedPath := strings.Replace(importPath, alias, target, 1)
			return strings.TrimPrefix(resolvedPath, "./")
		}
	}

	if strings.HasPrefix(importPath, "./") || strings.HasPrefix(importPath, "../") {
		return importPath
	}

	return importPath
}

func (da *DependencyAnalyzer) extractLayerAndSlice(importPath string) (string, string) {
	importPath = strings.TrimPrefix(importPath, "./")
	importPath = strings.TrimPrefix(importPath, "../")
	importPath = strings.TrimPrefix(importPath, "/")
	
	parts := strings.Split(importPath, "/")
	if len(parts) == 0 {
		return "", ""
	}
	
	for _, layer := range model.KnownLayers {
		if parts[0] == layer {
			if len(parts) > 1 {
				sliceName := parts[1]
				return layer, sliceName
			}
			return layer, layer
		}
	}
	
	if strings.Contains(importPath, "/") {
		pathParts := strings.Split(importPath, "/")
		
		for i, part := range pathParts {
			for _, layer := range model.KnownLayers {
				if part == layer && i < len(pathParts)-1 {
					sliceName := pathParts[i+1]
					return layer, sliceName
				} else if part == layer {
					return layer, layer
				}
			}
		}
	}
	
	return "", ""
}

func (da *DependencyAnalyzer) isAllowedCyclicalDependency(layerName string) bool {
	if da.config == nil || len(da.config.AllowedCyclicalDependencies) == 0 {
		return false
	}

	for _, allowed := range da.config.AllowedCyclicalDependencies {
		if layerName == allowed {
			return true
		}
	}

	return false
}

func (da *DependencyAnalyzer) isAllowedCyclicalSlice(layerName, sliceName string) bool {
	if da.config == nil || len(da.config.AllowedCyclicalDependencies) == 0 {
		return false
	}

	slicePath := layerName
	if sliceName != "" && sliceName != layerName {
		slicePath = layerName + "/" + sliceName
	}

	for _, allowed := range da.config.AllowedCyclicalDependencies {
		if slicePath == allowed {
			return true
		}
	}

	return false
}

func (da *DependencyAnalyzer) determineDependencyTypeWithSlices(fromLayer, fromSlice, toLayer, toSlice string) DependencyType {
	if fromLayer == "test" || toLayer == "test" {
		return DependencyTest
	}

	fromIndex, fromExists := da.layerIndices[fromLayer]
	toIndex, toExists := da.layerIndices[toLayer]
	
	if !fromExists || !toExists {
		return DependencyNormal
	}
	
	if fromLayer == toLayer {
		return DependencySameLayer
	}
	
	if fromIndex > toIndex {
		if da.isAllowedCyclicalSlice(fromLayer, fromSlice) || da.isAllowedCyclicalSlice(toLayer, toSlice) {
			return DependencyNormal
		}
		if da.isAllowedCyclicalDependency(fromLayer) || da.isAllowedCyclicalDependency(toLayer) {
			return DependencyNormal
		}
		return DependencyCyclical
	}
	
	return DependencyNormal
}

func (da *DependencyAnalyzer) GetDependenciesForLayer(layerName string) []Dependency {
	var result []Dependency
	for _, dep := range da.dependencies {
		if dep.FromLayer == layerName || dep.ToLayer == layerName {
			result = append(result, dep)
		}
	}
	return result
}

func (da *DependencyAnalyzer) GetDependenciesForSlice(layerName, sliceName string) []Dependency {
	var result []Dependency
	for _, dep := range da.dependencies {
		if (dep.FromLayer == layerName && dep.FromSlice == sliceName) || 
		   (dep.ToLayer == layerName && dep.ToSlice == sliceName) {
			result = append(result, dep)
		}
	}
	return result
}

func (da *DependencyAnalyzer) GetProblematicDependencies() []Dependency {
	var result []Dependency
	for _, dep := range da.dependencies {
		if dep.Type == DependencyCyclical {
			result = append(result, dep)
		}
	}
	return result
} 