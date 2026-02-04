package config

import (
	"os"
	"reflect"
	"testing"

	"github.com/DebroyeAntoine/flexigo/internal/types"
)

// ============================================================================
// Tests pour les couleurs
// ============================================================================

func TestLoadConfig_WithColors(t *testing.T) {
	yaml := `
default_color:
  r: 0
  g: 255
  b: 0
  a: 255
default_highlight_color:
  r: 255
  g: 255
  b: 0
  a: 255
blocks:
  - label: "Green Button"
    type: tts
    text: "Hello"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	// Vérifie que les couleurs par défaut sont appliquées
	if cfg.DefaultColor == nil {
		t.Fatal("expected DefaultColor to be set")
	}
	if cfg.DefaultColor.R != 0 || cfg.DefaultColor.G != 255 || cfg.DefaultColor.B != 0 {
		t.Errorf("unexpected default color: got RGB(%d,%d,%d)",
			cfg.DefaultColor.R, cfg.DefaultColor.G, cfg.DefaultColor.B)
	}

	// Vérifie que le bloc hérite de la couleur par défaut
	if cfg.Blocks[0].Color == nil {
		t.Fatal("expected block to have color set")
	}
	if cfg.Blocks[0].Color.G != 255 {
		t.Errorf("expected block to inherit green color, got RGB(%d,%d,%d)",
			cfg.Blocks[0].Color.R, cfg.Blocks[0].Color.G, cfg.Blocks[0].Color.B)
	}
}

func TestLoadConfig_CustomBlockColor(t *testing.T) {
	yaml := `
default_color:
  r: 255
  g: 0
  b: 0
  a: 255
blocks:
  - label: "Red Button"
    type: tts
    text: "Hello"
  - label: "Blue Button"
    type: tts
    text: "World"
    color:
      r: 0
      g: 0
      b: 255
      a: 255
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	// Premier bloc hérite du rouge par défaut
	if cfg.Blocks[0].Color.R != 255 || cfg.Blocks[0].Color.B != 0 {
		t.Errorf("first block should be red, got RGB(%d,%d,%d)",
			cfg.Blocks[0].Color.R, cfg.Blocks[0].Color.G, cfg.Blocks[0].Color.B)
	}

	// Deuxième bloc a sa propre couleur bleue
	if cfg.Blocks[1].Color.B != 255 || cfg.Blocks[1].Color.R != 0 {
		t.Errorf("second block should be blue, got RGB(%d,%d,%d)",
			cfg.Blocks[1].Color.R, cfg.Blocks[1].Color.G, cfg.Blocks[1].Color.B)
	}
}

func TestLoadConfig_NestedContainerColors(t *testing.T) {
	yaml := `
default_color:
  r: 255
  g: 0
  b: 0
  a: 255
blocks:
  - label: "Container"
    type: container
    children:
      - label: "Child 1"
        type: tts
        text: "Hello"
      - label: "Child 2"
        type: tts
        text: "World"
        color:
          r: 0
          g: 255
          b: 0
          a: 255
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	container := cfg.Blocks[0]

	// Container hérite de la couleur par défaut
	if container.Color.R != 255 {
		t.Errorf("container should inherit red color")
	}

	// Premier enfant hérite aussi
	if container.Children[0].Color.R != 255 {
		t.Errorf("first child should inherit red color")
	}

	// Deuxième enfant a sa propre couleur
	if container.Children[1].Color.G != 255 {
		t.Errorf("second child should be green")
	}
}

func TestLoadConfig_NoColorsUsesDefaults(t *testing.T) {
	yaml := `
blocks:
  - label: "Button"
    type: tts
    text: "Hello"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	// Devrait utiliser la couleur par défaut (rouge)
	if cfg.Blocks[0].Color == nil {
		t.Fatal("expected default color to be applied")
	}

	defaultColor := types.DefaultButtonColor()
	if cfg.Blocks[0].Color.R != defaultColor.R {
		t.Errorf("expected default red color, got RGB(%d,%d,%d)",
			cfg.Blocks[0].Color.R, cfg.Blocks[0].Color.G, cfg.Blocks[0].Color.B)
	}
}

// ============================================================================
// Tests pour les voix TTS
// ============================================================================

func TestLoadConfig_WithDefaultVoice(t *testing.T) {
	yaml := `
default_voice: "Microsoft David"
blocks:
  - label: "Say Hello"
    type: tts
    text: "Hello world"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	if cfg.DefaultVoice != "Microsoft David" {
		t.Errorf("expected default voice 'Microsoft David', got '%s'", cfg.DefaultVoice)
	}

	// Le bloc TTS devrait hériter de la voix par défaut
	if cfg.Blocks[0].Voice != "Microsoft David" {
		t.Errorf("expected block to inherit default voice, got '%s'", cfg.Blocks[0].Voice)
	}
}

func TestLoadConfig_CustomVoiceOverridesDefault(t *testing.T) {
	yaml := `
default_voice: "Microsoft David"
blocks:
  - label: "Say Hello"
    type: tts
    text: "Hello"
    voice: "Microsoft Zira"
  - label: "Say World"
    type: tts
    text: "World"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	// Premier bloc avec voix personnalisée
	if cfg.Blocks[0].Voice != "Microsoft Zira" {
		t.Errorf("expected custom voice 'Microsoft Zira', got '%s'", cfg.Blocks[0].Voice)
	}

	// Deuxième bloc hérite de la voix par défaut
	if cfg.Blocks[1].Voice != "Microsoft David" {
		t.Errorf("expected default voice, got '%s'", cfg.Blocks[1].Voice)
	}
}

func TestLoadConfig_VoiceOnlyForTTS(t *testing.T) {
	yaml := `
default_voice: "Microsoft David"
blocks:
  - label: "HTTP Call"
    type: http
    url: "http://example.com"
  - label: "Say Hello"
    type: tts
    text: "Hello"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	// Le bloc HTTP ne devrait pas avoir de voix
	if cfg.Blocks[0].Voice != "" {
		t.Errorf("HTTP block should not have voice, got '%s'", cfg.Blocks[0].Voice)
	}

	// Le bloc TTS devrait avoir la voix par défaut
	if cfg.Blocks[1].Voice != "Microsoft David" {
		t.Errorf("TTS block should have default voice, got '%s'", cfg.Blocks[1].Voice)
	}
}

func TestLoadConfig_NestedTTSVoices(t *testing.T) {
	yaml := `
default_voice: "Microsoft David"
blocks:
  - label: "Container"
    type: container
    children:
      - label: "Say 1"
        type: tts
        text: "One"
      - label: "Say 2"
        type: tts
        text: "Two"
        voice: "Microsoft Zira"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	container := cfg.Blocks[0]

	// Premier enfant hérite de la voix par défaut
	if container.Children[0].Voice != "Microsoft David" {
		t.Errorf("first child should inherit default voice, got '%s'",
			container.Children[0].Voice)
	}

	// Deuxième enfant a sa propre voix
	if container.Children[1].Voice != "Microsoft Zira" {
		t.Errorf("second child should have custom voice, got '%s'",
			container.Children[1].Voice)
	}
}

// ============================================================================
// Tests de validation
// ============================================================================

func TestLoadConfig_ComplexScenario(t *testing.T) {
	yaml := `
default_color:
  r: 100
  g: 100
  b: 100
  a: 255
default_highlight_color:
  r: 200
  g: 200
  b: 0
  a: 255
default_voice: "Microsoft David"
blocks:
  - label: "Main Container"
    type: container
    timer: 2000
    grid_width: 4
    grid_height: 4
    children:
      - label: "Red Button"
        type: tts
        text: "Red"
        voice: "Microsoft Zira"
        color:
          r: 255
          g: 0
          b: 0
          a: 255
        position:
          x: 0
          y: 0
      - label: "Default Button"
        type: tts
        text: "Default"
        position:
          x: 1
          y: 0
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	container := cfg.Blocks[0]

	// Vérification du timer hérité
	if container.Timer != 2000 {
		t.Errorf("expected timer 2000, got %d", container.Timer)
	}
	if container.Children[0].Timer != 2000 {
		t.Errorf("child should inherit timer 2000, got %d", container.Children[0].Timer)
	}

	// Vérification des couleurs
	redBtn := container.Children[0]
	if redBtn.Color.R != 255 || redBtn.Color.G != 0 {
		t.Errorf("red button should be red")
	}

	defaultBtn := container.Children[1]
	if defaultBtn.Color.R != 100 || defaultBtn.Color.G != 100 {
		t.Errorf("default button should have gray color")
	}

	// Vérification des voix
	if redBtn.Voice != "Microsoft Zira" {
		t.Errorf("red button should have custom voice")
	}
	if defaultBtn.Voice != "Microsoft David" {
		t.Errorf("default button should have default voice")
	}

	// Vérification des positions
	if redBtn.Position.X != 0 || redBtn.Position.Y != 0 {
		t.Errorf("red button position incorrect")
	}
	if defaultBtn.Position.X != 1 || defaultBtn.Position.Y != 0 {
		t.Errorf("default button position incorrect")
	}
}

func TestLoadConfig_DefaultTimerAndGroup(t *testing.T) {
	yaml := `
blocks:
  - label: "Container"
    type: container
    children:
      - label: "Child"
        type: tts
        text: "Hello"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("LoadConfig returned error: %v", err)
	}

	container := cfg.Blocks[0]
	if container.Timer != defaultTimer {
		t.Errorf("expected container default timer %d, got %d", defaultTimer, container.Timer)
	}
	if container.Children[0].Timer != defaultTimer {
		t.Errorf("expected child timer %d, got %d", defaultTimer, container.Children[0].Timer)
	}

	if container.Children[0].GroupMembership == nil || *container.Children[0].GroupMembership != 0 {
		t.Errorf("expected default group membership 0, got %v", container.Children[0].GroupMembership)
	}
}

// ============================================================================
// Tests pour ColorToImageColor
// ============================================================================

func TestColor_ToImageColor(t *testing.T) {
	color := types.Color{R: 255, G: 128, B: 64, A: 200}
	imgColor := color.ToImageColor()

	if imgColor == nil {
		t.Fatal("ToImageColor returned nil")
	}

	val := reflect.ValueOf(imgColor)
	if val.Kind() != reflect.Struct {
		t.Fatalf("expected struct color, got %s", val.Kind())
	}

	r := val.FieldByName("R")
	g := val.FieldByName("G")
	b := val.FieldByName("B")
	a := val.FieldByName("A")
	if !r.IsValid() || !g.IsValid() || !b.IsValid() || !a.IsValid() {
		t.Fatal("ToImageColor did not expose RGBA fields")
	}

	if r.Uint() != 255 || g.Uint() != 128 || b.Uint() != 64 || a.Uint() != 200 {
		t.Errorf("color values incorrect: got RGBA(%d,%d,%d,%d)",
			r.Uint(), g.Uint(), b.Uint(), a.Uint())
	}
}

// ============================================================================
// Tests pour l'IR config
// ============================================================================

func TestValidateIRConfig_SerialRequiresPort(t *testing.T) {
	cfg := &types.Config{
		IRBackend: "serial",
	}

	err := ValidateIRConfig(cfg)
	if err == nil {
		t.Fatal("expected error when IRSerialPort is empty for serial backend")
	}
}

func TestValidateIRConfig_DefaultBaudRate(t *testing.T) {
	cfg := &types.Config{
		IRBackend:    "serial",
		IRSerialPort: "/dev/ttyUSB0",
		IRBaudRate:   0,
	}

	err := ValidateIRConfig(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.IRBaudRate != 9600 {
		t.Errorf("expected baud rate 9600, got %d", cfg.IRBaudRate)
	}
}

// ============================================================================
// Helpers
// ============================================================================

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()

	file, err := os.CreateTemp("", "flexigo-config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		t.Fatalf("failed to write temp YAML: %v", err)
	}

	return file.Name()
}
