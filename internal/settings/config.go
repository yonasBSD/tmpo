package settings

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// IMPORTANT: When adding new fields to this struct, also update configTemplate below.
type Config struct {
	ProjectName string `yaml:"project_name"`
	HourlyRate float64 `yaml:"hourly_rate,omitempty"`
	Description string `yaml:"description,omitempty"`
	ExportPath  string `yaml:"export_path,omitempty"`
}

// IMPORTANT: When adding new fields to Config, update this template.
const configTemplate = `# tmpo project configuration
# This file configures time tracking settings for this project

# Project name (used to identify time entries)
project_name: %s

# [OPTIONAL] Hourly rate for billing calculations (set to 0 to disable)
hourly_rate: %.2f

# [OPTIONAL] Description for this project
description: "%s"

# [OPTIONAL] Default export path for this project (overrides global export path)
export_path: "%s"
`

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func Create(projectName string, hourlyRate float64) error {
	config := &Config{
		ProjectName: projectName,
		HourlyRate: hourlyRate,
	}

	tmporc := filepath.Join(".", ".tmporc")
	if _, err := os.Stat(tmporc); err == nil {
		return fmt.Errorf(".tmporc already exists")
	}

	return config.Save(tmporc)
}

func CreateWithTemplate(projectName string, hourlyRate float64, description string, exportPath string) error {
	tmporc := filepath.Join(".", ".tmporc")
	if _, err := os.Stat(tmporc); err == nil {
		return fmt.Errorf(".tmporc already exists")
	}

	content := fmt.Sprintf(configTemplate, projectName, hourlyRate, description, exportPath)

	if err := os.WriteFile(tmporc, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

func FindAndLoad() (*Config, string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}

	for {
		tmporc := filepath.Join(dir, ".tmporc")
		if _, err := os.Stat(tmporc); err == nil {
			config, err := Load(tmporc)

			return config, tmporc, err
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}

		dir = parent
	}

	return nil, "", fmt.Errorf(".tmporc not found")
}
