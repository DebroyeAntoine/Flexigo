package config

import (
	"os"
	"testing"
)

// For each test case, we need to create a config file
func writeTempYAML(t *testing.T, content string) string {
	tmpFile, err := os.CreateTemp("", "config_*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	if _, err := tmpFile.Write([]byte(content)); err != nil {
		t.Fatalf("failed to write to temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("failed to close temp file: %v", err)
	}

	return tmpFile.Name()
}

func TestLoadConfig_Basic(t *testing.T) {
	yaml := `
blocks:
  - label: "Main Menu"
    type: container
    children:
      - label: "Say Hello"
        type: tts
        text: "Hello world!"
`
	file := writeTempYAML(t, yaml)

	cfg, err := LoadConfig(file)
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

func TestLoadConfig_ErrorCases(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
	}{
		{
			name: "Invalid YAML",
			yaml: `
blocks:
  - label: "Bad Block"
    type: container
    children:
      - label: "Oops"
        type: http
        url: "http://example.com
`,
			wantErr: true,
		},
		{
			name:    "Empty file",
			yaml:    ``,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := writeTempYAML(t, tt.yaml)
			cfg, err := LoadConfig(file)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.name == "Empty file" && len(cfg.Blocks) != 0 {
					t.Fatalf("expected 0 blocks, got %d", len(cfg.Blocks))
				}
			}
		})
	}
}

func TestLoadConfig_UnknownType(t *testing.T) {
	yaml := `
blocks:
  - label: "Unknown Action"
    type: teleport
`
	file := writeTempYAML(t, yaml)

	cfg, err := LoadConfig(file)
	if err != nil {
		t.Fatalf("unexpected error loading unknown type: %v", err)
	}

	if cfg.Blocks[0].Type != "teleport" {
		t.Errorf("expected block type to be 'teleport', got '%s'", cfg.Blocks[0].Type)
	}
}

func TestLoadConfig_MissingFields(t *testing.T) {
	tests := []struct {
		name       string
		yaml       string
		blockType  string
		expectText string
		expectURL  string
		expectCmd  string
	}{
		{
			name: "TTS block missing text",
			yaml: `
blocks:
  - label: "Say something"
    type: tts
`,
			blockType: "tts",
		},
		{
			name: "HTTP block missing URL",
			yaml: `
blocks:
  - label: "Send request"
    type: http
    method: POST
`,
			blockType: "http",
		},
		{
			name: "Exec block missing command",
			yaml: `
blocks:
  - label: "Run script"
    type: exec
    args: ["--version"]
`,
			blockType: "exec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := writeTempYAML(t, tt.yaml)
			cfg, err := LoadConfig(file)
			if err != nil {
				t.Fatalf("unexpected error loading config: %v", err)
			}
			block := cfg.Blocks[0]
			if block.Type != tt.blockType {
				t.Errorf("expected block type '%s', got '%s'", tt.blockType, block.Type)
			}
			if block.Text != tt.expectText {
				t.Errorf("expected text '%s', got '%s'", tt.expectText, block.Text)
			}
			if block.URL != tt.expectURL {
				t.Errorf("expected URL '%s', got '%s'", tt.expectURL, block.URL)
			}
			if block.Command != tt.expectCmd {
				t.Errorf("expected command '%s', got '%s'", tt.expectCmd, block.Command)
			}
		})
	}
}
