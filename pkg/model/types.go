package model

import "fsd-crawler/pkg/config"

type FSDLayer struct {
	Name   string
	Slices []*FSDSlice
}

type FSDSlice struct {
	Name     string
	Segments []*FSDSegment
}

type FSDSegment struct {
	Name  string
	Files []string
}

type ProjectStructure struct {
	Layers []*FSDLayer
	Dependencies []interface{}
}

var KnownLayers = []string{"app", "processes", "pages", "widgets", "features", "entities", "shared"}

var KnownSegments = []string{"ui", "api", "model", "lib", "config"}

func UpdateFromConfig(cfg *config.Config) {
	if cfg == nil {
		return
	}

	if len(cfg.CustomLayers) > 0 {
		KnownLayers = cfg.CustomLayers
	}
} 