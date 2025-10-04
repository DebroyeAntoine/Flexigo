package config

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"

	"github.com/DebroyeAntoine/flexigo/internal/types"
)

const defaultTimer = 1000 // in ms

func applyDefaultTimer(actions []types.Action, defaultTimer int) {
	for i := range actions {
		if actions[i].Timer == 0 {
			actions[i].Timer = defaultTimer
		}

		if actions[i].Type == "container" {
			// For the inheritance, give to children the parent timer as default
			applyDefaultTimer(actions[i].Children, actions[i].Timer)
		}
	}
}

func UniformizeTimer(cfg *types.Config) {
	if len(cfg.Blocks) == 0 {
		return
	}

	for i := range cfg.Blocks {
		if cfg.Blocks[i].Timer == 0 {
			cfg.Blocks[i].Timer = defaultTimer
		}
		if cfg.Blocks[i].Type == "container" {
			applyDefaultTimer(cfg.Blocks[i].Children, cfg.Blocks[i].Timer)
		}
	}
}

func ApplyDefaultGroup(actions []types.Action, defaultGroup int) {
	for i := range actions {
		if actions[i].GroupMembership == nil {
			actions[i].GroupMembership = &defaultGroup
		}
	}
}

func CreateDefaultGroup(cfg *types.Config) {
	for i := range cfg.Blocks {
		if cfg.Blocks[i].Type == "container" {
			ApplyDefaultGroup(cfg.Blocks[i].Children, 0)
		}
	}
}

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

	CreateDefaultGroup(&cfg)

	UniformizeTimer(&cfg)

	return &cfg, nil
}
