package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"github.com/DebroyeAntoine/flexigo/internal/types"
)

func LoadConfig(path string) (*types.Config, error) {
	_ = godotenv.Load(".env")

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, err
	}

	var cfg types.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
