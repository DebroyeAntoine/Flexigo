package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Basic(t *testing.T) {
	// YAML de test
	yamlData := `
blocks:
  - label: "Menu Principal"
    type: container
    children:
      - label: "Dire bonjour"
        type: tts
        text: "Bonjour !"
`

	tmpFile, err := os.CreateTemp("", "test_config_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadConfig returned an error: %v", err)
	}

	if len(cfg.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(cfg.Blocks))
	}

	root := cfg.Blocks[0]
	if root.Type != "container" {
		t.Errorf("expected block type 'container', got '%s'", root.Type)
	}
	if len(root.Children) != 1 {
		t.Fatalf("expected 1 child block, got %d", len(root.Children))
	}
	child := root.Children[0]
	if child.Type != "tts" || child.Text != "Bonjour !" {
		t.Errorf("unexpected child content: %+v", child)
	}
}
