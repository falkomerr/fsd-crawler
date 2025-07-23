package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	SrcDir                   string            `yaml:"srcDir"`
	OutputDir                string            `yaml:"outputDir"`
	OutputFormats            []string          `yaml:"outputFormats"`
	ExcludeDirs              []string          `yaml:"excludeDirs"`
	CustomLayers             []string          `yaml:"customLayers"`
	HTMLTemplatePath         string            `yaml:"htmlTemplatePath"`
	Aliases                  map[string]string `yaml:"aliases"`
	ServeHTML                bool              `yaml:"serveHTML"`
	Port                     int               `yaml:"port"`
	AllowedCyclicalDependencies []string       `yaml:"allowedCyclicalDependencies"`
}

var DefaultConfig = Config{
	SrcDir:        ".",
	OutputDir:     "./dist",
	OutputFormats: []string{"html"},
	ExcludeDirs:   []string{"node_modules", ".git", "dist", "build"},
	Aliases: map[string]string{
		"@": "src",
	},
	ServeHTML: true,
	Port: 3123,
	AllowedCyclicalDependencies: []string{},
}

func FindAndLoadConfig() (*Config, error) {
	configPaths := []string{
		"fsd-crawler.yml",
		"fsd-crawler.yaml",
		".fsd-crawler.yml",
		".fsd-crawler.yaml",
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return LoadConfig(path)
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("не удалось получить текущую директорию: %v", err)
	}

	for {
		for _, configName := range []string{"fsd-crawler.yml", "fsd-crawler.yaml", ".fsd-crawler.yml", ".fsd-crawler.yaml"} {
			configPath := filepath.Join(dir, configName)
			if _, err := os.Stat(configPath); err == nil {
				return LoadConfig(configPath)
			}
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break
		}
		dir = parentDir
	}

	return &DefaultConfig, nil
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл конфигурации %s: %v", path, err)
	}

	config := DefaultConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("не удалось распарсить файл конфигурации %s: %v", path, err)
	}

	return &config, nil
} 