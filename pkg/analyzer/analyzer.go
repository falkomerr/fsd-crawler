package analyzer

import (
	"os"
	"path/filepath"
	"strings"

	"fsd-crawler/pkg/config"
	"fsd-crawler/pkg/dependencies"
	"fsd-crawler/pkg/model"
)

func AnalyzeProject(cfg *config.Config) *model.ProjectStructure {
	rootDir := cfg.SrcDir
	structure := &model.ProjectStructure{
		Layers: []*model.FSDLayer{},
	}

	model.UpdateFromConfig(cfg)

	for _, layerName := range model.KnownLayers {
		layerPath := filepath.Join(rootDir, layerName)
		
		if _, err := os.Stat(layerPath); os.IsNotExist(err) {
			continue
		}
		
		layer := &model.FSDLayer{
			Name:   layerName,
			Slices: []*model.FSDSlice{},
		}
		
		structure.Layers = append(structure.Layers, layer)
		
		analyzeLayer(layer, layerPath)
	}
	
	depAnalyzer := dependencies.NewDependencyAnalyzer(structure, rootDir, cfg)
	deps := depAnalyzer.AnalyzeDependencies()
	
	structure.Dependencies = make([]interface{}, len(deps))
	for i, dep := range deps {
		structure.Dependencies[i] = dep
	}
	
	return structure
}

func analyzeLayer(layer *model.FSDLayer, layerPath string) {
	entries, err := os.ReadDir(layerPath)
	if err != nil {
		return
	}
	
	hasFiles := false
	for _, entry := range entries {
		if !entry.IsDir() && isSourceFile(entry.Name()) {
			hasFiles = true
			break
		}
	}
	
	if hasFiles {
		slice := &model.FSDSlice{
			Name:     layer.Name,
			Segments: []*model.FSDSegment{},
		}
		segment := &model.FSDSegment{
			Name:  "root",
			Files: []string{},
		}
		
		for _, entry := range entries {
			if !entry.IsDir() && isSourceFile(entry.Name()) {
				segment.Files = append(segment.Files, entry.Name())
			}
		}
		
		slice.Segments = append(slice.Segments, segment)
		layer.Slices = append(layer.Slices, slice)
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			slicePath := filepath.Join(layerPath, entry.Name())
			slice := &model.FSDSlice{
				Name:     entry.Name(),
				Segments: []*model.FSDSegment{},
			}
			
			sliceEntries, err := os.ReadDir(slicePath)
			if err != nil {
				continue
			}
			
			hasRootFiles := false
			for _, sliceEntry := range sliceEntries {
				if !sliceEntry.IsDir() && isSourceFile(sliceEntry.Name()) {
					hasRootFiles = true
					break
				}
			}
			
			if hasRootFiles {
				segment := &model.FSDSegment{
					Name:  "root",
					Files: []string{},
				}
				
				for _, sliceEntry := range sliceEntries {
					if !sliceEntry.IsDir() && isSourceFile(sliceEntry.Name()) {
						segment.Files = append(segment.Files, sliceEntry.Name())
					}
				}
				
				slice.Segments = append(slice.Segments, segment)
			}
			
			analyzeSlice(slice, slicePath)
			
			if len(slice.Segments) > 0 {
				layer.Slices = append(layer.Slices, slice)
			}
		}
	}
}

func analyzeSlice(slice *model.FSDSlice, slicePath string) {
	entries, err := os.ReadDir(slicePath)
	if err != nil {
		return
	}
	
	for _, entry := range entries {
		if entry.IsDir() {
			segmentName := entry.Name()
			segmentPath := filepath.Join(slicePath, segmentName)
			
			isKnownSegment := false
			for _, knownSegment := range model.KnownSegments {
				if segmentName == knownSegment {
					isKnownSegment = true
					break
				}
			}
			
			if !isKnownSegment {
				continue
			}
			
			segment := &model.FSDSegment{
				Name:  segmentName,
				Files: []string{},
			}
			
			segmentEntries, err := os.ReadDir(segmentPath)
			if err != nil {
				continue
			}
			
			for _, segmentEntry := range segmentEntries {
				if !segmentEntry.IsDir() && isSourceFile(segmentEntry.Name()) {
					segment.Files = append(segment.Files, segmentEntry.Name())
				}
			}
			
			if len(segment.Files) > 0 {
				slice.Segments = append(slice.Segments, segment)
			}
		}
	}
}

func isSourceFile(fileName string) bool {
	ext := strings.ToLower(filepath.Ext(fileName))
	return ext == ".js" || ext == ".jsx" || ext == ".ts" || ext == ".tsx" || ext == ".vue"
} 