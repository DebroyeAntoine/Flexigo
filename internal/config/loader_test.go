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

func TestLoadConfig_InvalidYAML(t *testing.T) {
	invalidYAML := `
blocks:
  - label: "Bad Block"
    type: container
    children:
      - label: "Oops"
        type: http
        url: "http://example.com
`

	tmpFile, err := os.CreateTemp("", "invalid_config_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(invalidYAML)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	_, err = LoadConfig(tmpFile.Name())
	if err == nil {
		t.Fatal("expected LoadConfig to fail on invalid YAML, but it succeeded")
	}
}

func TestLoadConfig_EmptyFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "empty_config_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("expected no error on empty file, got: %v", err)
	}
	if len(cfg.Blocks) != 0 {
		t.Fatalf("expected 0 blocks, got %d", len(cfg.Blocks))
	}
}

func TestLoadConfig_UnknownType(t *testing.T) {
	yamlData := `
blocks:
  - label: "Unknown Action"
    type: teleport
`

	tmpFile, err := os.CreateTemp("", "unknown_type_config_*.yaml")
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
		t.Fatalf("unexpected error loading unknown type: %v", err)
	}

	if cfg.Blocks[0].Type != "teleport" {
		t.Errorf("expected block type to be 'teleport', got '%s'", cfg.Blocks[0].Type)
	}
}

func TestLoadConfig_HTTPBlockMissingURL(t *testing.T) {
	yamlData := `
blocks:
  - label: "Send request"
    type: http
    method: POST
`

	tmpFile, err := os.CreateTemp("", "http_missing_url_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(yamlData)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	block := cfg.Blocks[0]
	if block.Type != "http" || block.URL != "" {
		t.Errorf("expected empty URL for http block, got '%s'", block.URL)
	}
}

func TestLoadConfig_TTSBlockMissingText(t *testing.T) {
	yamlData := `
blocks:
  - label: "Say something"
    type: tts
`

	tmpFile, err := os.CreateTemp("", "tts_missing_text_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(yamlData)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	block := cfg.Blocks[0]
	if block.Type != "tts" || block.Text != "" {
		t.Errorf("expected empty text for tts block, got '%s'", block.Text)
	}
}

func TestLoadConfig_ExecBlockMissingCommand(t *testing.T) {
	yamlData := `
blocks:
  - label: "Run script"
    type: exec
    args: ["--version"]
`

	tmpFile, err := os.CreateTemp("", "exec_missing_command_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString(yamlData)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	block := cfg.Blocks[0]
	if block.Type != "exec" || block.Command != "" {
		t.Errorf("expected empty command for exec block, got '%s'", block.Command)
	}
}
