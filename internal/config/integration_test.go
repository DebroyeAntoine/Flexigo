package config

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpClient "github.com/DebroyeAntoine/flexigo/internal/http"
	"github.com/DebroyeAntoine/flexigo/internal/ir"
	"github.com/DebroyeAntoine/flexigo/internal/orchestration"
)

type recordingTTS struct {
	calls []ttsCall
}

type ttsCall struct {
	text  string
	voice string
}

func (r *recordingTTS) Say(text string) error {
	r.calls = append(r.calls, ttsCall{text: text, voice: ""})
	return nil
}

func (r *recordingTTS) SayWithVoice(text string, voice string) error {
	r.calls = append(r.calls, ttsCall{text: text, voice: voice})
	return nil
}

func (r *recordingTTS) ListVoices() ([]string, error) {
	return []string{"Voice1"}, nil
}

func TestIntegration_ConfigToOrchestration(t *testing.T) {
	yaml := `
default_voice: "Microsoft David"
blocks:
  - label: "Say Default"
    type: tts
    text: "Hello"
  - label: "Say Custom"
    type: tts
    text: "World"
    voice: "Microsoft Zira"
  - label: "HTTP Call"
    type: http
    url: "http://placeholder"
    body: '{"ping":"pong"}'
  - label: "IR Power"
    type: ir
    ir_device: "tv"
    ir_command: "power"
    ir_protocol: "NEC"
    ir_code: "0x20DF10EF"
    ir_repeat: 2
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("loadFromFile returned error: %v", err)
	}

	requestReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tts := &recordingTTS{}
	mockIR := ir.NewMockIRSender()
	o := orchestration.Orchestration{
		TTS:  tts,
		HTTP: httpClient.NewHTTPClient(),
		Cfg:  cfg,
		IR:   mockIR,
	}

	for i := range cfg.Blocks {
		if cfg.Blocks[i].Type == "http" {
			cfg.Blocks[i].URL = server.URL
		}
		if err := o.ExecuteTTSAction(cfg.Blocks[i]); err != nil {
			t.Fatalf("ExecuteTTSAction error: %v", err)
		}
		if err := o.ExecuteHTTPAction(cfg.Blocks[i]); err != nil {
			t.Fatalf("ExecuteHTTPAction error: %v", err)
		}
		if err := o.ExecuteIRAction(cfg.Blocks[i]); err != nil {
			t.Fatalf("ExecuteIRAction error: %v", err)
		}
	}

	if len(tts.calls) != 2 {
		t.Fatalf("expected 2 TTS calls, got %d", len(tts.calls))
	}
	if tts.calls[0].voice != "Microsoft David" {
		t.Errorf("expected default voice 'Microsoft David', got '%s'", tts.calls[0].voice)
	}
	if tts.calls[1].voice != "Microsoft Zira" {
		t.Errorf("expected custom voice 'Microsoft Zira', got '%s'", tts.calls[1].voice)
	}

	if !requestReceived {
		t.Error("expected HTTP request to be made")
	}

	lastCmd, err := mockIR.GetLastCommand()
	if err != nil {
		t.Fatalf("GetLastCommand failed: %v", err)
	}
	if lastCmd.Device != "tv" || lastCmd.Command != "power" || lastCmd.Protocol != "NEC" || lastCmd.Code != "0x20DF10EF" || lastCmd.Repeat != 2 {
		t.Errorf("IR command mismatch: %+v", lastCmd)
	}
}

func TestIntegration_HTTPErrorPropagates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	yaml := `
blocks:
  - label: "Bad HTTP"
    type: http
    url: "http://placeholder"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("loadFromFile returned error: %v", err)
	}

	o := orchestration.Orchestration{
		TTS:  &recordingTTS{},
		HTTP: httpClient.NewHTTPClient(),
		Cfg:  cfg,
		IR:   nil,
	}

	cfg.Blocks[0].URL = server.URL
	err = o.ExecuteHTTPAction(cfg.Blocks[0])
	if err == nil {
		t.Fatal("expected error for HTTP 404, got nil")
	}
}

func TestIntegration_DefaultsAppliedToNestedActions(t *testing.T) {
	yaml := `
default_color:
  r: 10
  g: 20
  b: 30
  a: 255
default_highlight_color:
  r: 40
  g: 50
  b: 60
  a: 255
default_background:
  r: 1
  g: 2
  b: 3
  a: 255
default_voice: "Microsoft David"
blocks:
  - label: "Container"
    type: container
    children:
      - label: "Say"
        type: tts
        text: "Hello"
      - label: "HTTP"
        type: http
        url: "http://example.com"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("loadFromFile returned error: %v", err)
	}

	container := cfg.Blocks[0]
	if container.Timer != defaultTimer {
		t.Errorf("expected container timer %d, got %d", defaultTimer, container.Timer)
	}

	if container.Color == nil || container.HighlightColor == nil || container.Background == nil {
		t.Fatal("expected container colors to be set")
	}
	if container.Color.R != 10 || container.HighlightColor.G != 50 || container.Background.B != 3 {
		t.Errorf("unexpected container colors: %+v %+v %+v", container.Color, container.HighlightColor, container.Background)
	}

	if len(container.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(container.Children))
	}

	ttsChild := container.Children[0]
	if ttsChild.Timer != defaultTimer {
		t.Errorf("expected child timer %d, got %d", defaultTimer, ttsChild.Timer)
	}
	if ttsChild.Voice != "Microsoft David" {
		t.Errorf("expected default voice on TTS, got %q", ttsChild.Voice)
	}
	if ttsChild.GroupMembership == nil || *ttsChild.GroupMembership != 0 {
		t.Errorf("expected default group membership 0, got %v", ttsChild.GroupMembership)
	}
	if ttsChild.Color == nil || ttsChild.HighlightColor == nil || ttsChild.Background == nil {
		t.Fatal("expected child colors to be set")
	}

	httpChild := container.Children[1]
	if httpChild.Voice != "" {
		t.Errorf("expected no voice on HTTP action, got %q", httpChild.Voice)
	}
}

func TestIntegration_IRValidationSerialRequiresPort(t *testing.T) {
	yaml := `
ir_backend: "serial"
blocks:
  - label: "Say"
    type: tts
    text: "Hello"
`
	file := writeTempYAML(t, yaml)
	_, err := loadFromFile(file)
	if err == nil {
		t.Fatal("expected error for serial backend without port, got nil")
	}
}

func TestIntegration_IRValidationDefaultBaudRate(t *testing.T) {
	yaml := `
ir_backend: "serial"
ir_serial_port: "/dev/ttyUSB0"
blocks:
  - label: "Say"
    type: tts
    text: "Hello"
`
	file := writeTempYAML(t, yaml)
	cfg, err := loadFromFile(file)
	if err != nil {
		t.Fatalf("loadFromFile returned error: %v", err)
	}
	if cfg.IRBaudRate != 9600 {
		t.Errorf("expected baud rate 9600, got %d", cfg.IRBaudRate)
	}
}
