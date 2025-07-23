package model

import (
	"testing"
)

func TestFSDStructureTypes(t *testing.T) {
	// Проверяем создание слоя
	layer := &FSDLayer{
		Name: "app",
	}

	if layer.Name != "app" {
		t.Errorf("FSDLayer.Name = %s; want app", layer.Name)
	}

	// Проверяем создание слайса
	slice := &FSDSlice{
		Name: "user",
	}

	if slice.Name != "user" {
		t.Errorf("FSDSlice.Name = %s; want user", slice.Name)
	}

	// Проверяем создание сегмента
	segment := &FSDSegment{
		Name:  "ui",
		Files: []string{"component.tsx"},
	}

	if segment.Name != "ui" {
		t.Errorf("FSDSegment.Name = %s; want ui", segment.Name)
	}

	if len(segment.Files) != 1 || segment.Files[0] != "component.tsx" {
		t.Errorf("FSDSegment.Files = %v; want [component.tsx]", segment.Files)
	}

	// Проверяем наличие известных слоев
	expectedLayers := []string{"app", "processes", "pages", "widgets", "features", "entities", "shared"}
	if len(KnownLayers) != len(expectedLayers) {
		t.Errorf("KnownLayers length = %d; want %d", len(KnownLayers), len(expectedLayers))
	}

	for i, layer := range KnownLayers {
		if layer != expectedLayers[i] {
			t.Errorf("KnownLayers[%d] = %s; want %s", i, layer, expectedLayers[i])
		}
	}

	// Проверяем наличие известных сегментов
	expectedSegments := []string{"ui", "api", "model", "lib", "config"}
	if len(KnownSegments) != len(expectedSegments) {
		t.Errorf("KnownSegments length = %d; want %d", len(KnownSegments), len(expectedSegments))
	}

	for i, segment := range KnownSegments {
		if segment != expectedSegments[i] {
			t.Errorf("KnownSegments[%d] = %s; want %s", i, segment, expectedSegments[i])
		}
	}
} 