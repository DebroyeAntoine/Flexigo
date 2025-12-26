package config

import (
	"fmt"
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

func ValidateIRConfig(cfg *types.Config) error {
	if cfg.IRBackend == "" {
		return nil // Pas d'IR configuré, c'est valide
	}

	if cfg.IRBackend == "serial" {
		if cfg.IRSerialPort == "" {
			return fmt.Errorf("IRBackend est 'serial' mais IRSerialPort est vide dans la configuration")
		}
		if cfg.IRBaudRate <= 0 {
			cfg.IRBaudRate = 9600 // Valeur par défaut si manquant
		}
	}
	return nil
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

// applyDefaultColors applique récursivement les couleurs par défaut
func applyDefaultColors(actions []types.Action, defaultColor, defaultHighlight *types.Color) {
	for i := range actions {
		// Applique la couleur par défaut si non définie
		if actions[i].Color == nil {
			actions[i].Color = defaultColor
		}

		// Applique la couleur de highlight par défaut si non définie
		if actions[i].HighlightColor == nil {
			actions[i].HighlightColor = defaultHighlight
		}

		// Applique récursivement aux enfants
		if actions[i].Type == "container" && len(actions[i].Children) > 0 {
			applyDefaultColors(actions[i].Children, defaultColor, defaultHighlight)
		}
	}
}

// UniformizeColors applique les couleurs par défaut à tous les blocs
func UniformizeColors(cfg *types.Config) {
	// Définit les couleurs par défaut au niveau de la config si non définies
	if cfg.DefaultColor == nil {
		cfg.DefaultColor = types.DefaultButtonColor()
	}
	if cfg.HighlightColor == nil {
		cfg.HighlightColor = types.DefaultHighlightColor()
	}

	// Applique aux blocs
	applyDefaultColors(cfg.Blocks, cfg.DefaultColor, cfg.HighlightColor)
}

// applyDefaultVoice applique récursivement la voix par défaut aux actions TTS
func applyDefaultVoice(actions []types.Action, defaultVoice string) {
	for i := range actions {
		// Applique la voix par défaut uniquement aux actions TTS sans voix définie
		if actions[i].Type == "tts" && actions[i].Voice == "" {
			actions[i].Voice = defaultVoice
		}

		// Applique récursivement aux enfants
		if actions[i].Type == "container" && len(actions[i].Children) > 0 {
			applyDefaultVoice(actions[i].Children, defaultVoice)
		}
	}
}

// UniformizeVoice applique la voix par défaut à toutes les actions TTS
func UniformizeVoice(cfg *types.Config) {
	if cfg.DefaultVoice == "" {
		return // Pas de voix par défaut définie
	}

	applyDefaultVoice(cfg.Blocks, cfg.DefaultVoice)
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

	// Applique les valeurs par défaut dans l'ordre
	CreateDefaultGroup(&cfg)
	UniformizeTimer(&cfg)
	UniformizeColors(&cfg)
	UniformizeVoice(&cfg)
	if err := ValidateIRConfig(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
