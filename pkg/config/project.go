package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const ProjectConfigFile = ".envv/config.yaml"

// ProjectConfig represents project-specific configuration
type ProjectConfig struct {
	OrganizationID   string `yaml:"organization_id"`
	OrganizationName string `yaml:"organization_name"`
	ProjectID        string `yaml:"project_id"`
	ProjectName      string `yaml:"project_name"`
	DefaultEnv       string `yaml:"default_environment"`
}

// LoadProjectConfig reads project config from current directory
func LoadProjectConfig() (*ProjectConfig, error) {
	data, err := os.ReadFile(ProjectConfigFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("not in envv project. Run 'envv init'")
		}
		return nil, fmt.Errorf("failed to read project config: %w", err)
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid project config: %w", err)
	}

	return &cfg, nil
}

// SaveProjectConfig writes project config to current directory
func SaveProjectConfig(cfg *ProjectConfig) error {
	// Create .envv directory
	if err := os.MkdirAll(".envv", 0755); err != nil {
		return fmt.Errorf("failed to create .envv directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal project config: %w", err)
	}

	if err := os.WriteFile(ProjectConfigFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	return nil
}

// ProjectConfigExists checks if project config exists
func ProjectConfigExists() bool {
	_, err := os.Stat(ProjectConfigFile)
	return err == nil
}
