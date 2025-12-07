package config

import (
	"fmt"
	"os"
	"path/filepath"

	"go.yaml.in/yaml/v3"
)

// Config represents the project's configuration as loaded from a YAML file.
// It contains identifying and billing information along with an optional description.
//
// ProjectName is the human-readable name of the project.
// HourlyRate is the billable hourly rate for the project; when zero it will be omitted from YAML.
// Description is an optional free-form description of the project; when empty it will be omitted from YAML.
type Config struct {
	ProjectName string `yaml:"project_name"`
	HourlyRate float64 `yaml:"hourly_rate,omitempty"`
	Description string `yaml:"description,omitempty"`
}

// Load reads a YAML configuration file from the provided path and unmarshals it into a Config.
// It returns a pointer to the populated Config on success. If the file cannot be read or the
// contents cannot be parsed as YAML, Load returns a wrapped error describing the failure.
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

// Save marshals the Config into YAML and writes it to the provided filesystem path.
// The configuration is encoded using yaml.Marshal and written with file mode 0644.
// If a file already exists at path it will be overwritten. An error is returned
// if marshaling fails or if writing the file to disk is unsuccessful.
func (c* Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// Create creates a new Config populated with projectName and hourlyRate and writes it
// to a ".tmporc" file in the current working directory ("./.tmporc").
// If a ".tmporc" file already exists, Create returns an error and does not overwrite it.
// Any error encountered while saving the configuration is returned to the caller.
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

// FindAndLoad searches upward from the current working directory for a file named
// ".tmporc". Starting at os.Getwd(), it ascends parent directories until it either
// finds the file or reaches the filesystem root.
//
// If a ".tmporc" file is found, FindAndLoad calls Load(path) and returns the resulting
// *Config, the absolute path to the discovered file, and any error returned by Load.
// If Load returns an error the returned *Config may be nil.
//
// If the current working directory cannot be determined, or if no ".tmporc" is found
// before reaching the root, FindAndLoad returns (nil, "", err) describing the failure.
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
