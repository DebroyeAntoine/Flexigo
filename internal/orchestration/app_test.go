package orchestration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpClient "github.com/DebroyeAntoine/flexigo/internal/http"
	"github.com/DebroyeAntoine/flexigo/internal/ir"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

type fakeTTS struct {
	calledText  string
	calledVoice string
	sayCount    int
}

func (f *fakeTTS) Say(text string) error {
	f.calledText = text
	f.calledVoice = ""
	f.sayCount++
	return nil
}

func (f *fakeTTS) SayWithVoice(text string, voice string) error {
	f.calledText = text
	f.calledVoice = voice
	f.sayCount++
	return nil
}

func (f *fakeTTS) ListVoices() ([]string, error) {
	return []string{"Voice1", "Voice2"}, nil
}

func TestOrchestrationSay(t *testing.T) {
	mockTTS := &fakeTTS{}
	cfg := &types.Config{}

	o := Orchestration{TTS: mockTTS, HTTP: httpClient.NewHTTPClient(), Cfg: cfg}

	err := o.Say("Hello test")
	if err != nil {
		t.Fatalf("Say returned unexpected error: %v", err)
	}

	if mockTTS.calledText != "Hello test" {
		t.Errorf("Expected TTS.Say to be called with 'Hello test', got '%s'", mockTTS.calledText)
	}

	if mockTTS.sayCount != 1 {
		t.Errorf("Expected Say to be called once, got %d", mockTTS.sayCount)
	}
}

func TestOrchestrationSayWithVoice(t *testing.T) {
	mockTTS := &fakeTTS{}
	cfg := &types.Config{}

	o := Orchestration{TTS: mockTTS, HTTP: httpClient.NewHTTPClient(), Cfg: cfg}

	tests := []struct {
		name  string
		text  string
		voice string
	}{
		{
			name:  "Say with Microsoft David",
			text:  "Hello",
			voice: "Microsoft David",
		},
		{
			name:  "Say with Microsoft Zira",
			text:  "World",
			voice: "Microsoft Zira",
		},
		{
			name:  "Say with empty voice",
			text:  "Test",
			voice: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTTS.calledText = ""
			mockTTS.calledVoice = ""

			err := o.SayWithVoice(tt.text, tt.voice)
			if err != nil {
				t.Fatalf("SayWithVoice returned unexpected error: %v", err)
			}

			if mockTTS.calledText != tt.text {
				t.Errorf("Expected text '%s', got '%s'", tt.text, mockTTS.calledText)
			}

			if mockTTS.calledVoice != tt.voice {
				t.Errorf("Expected voice '%s', got '%s'", tt.voice, mockTTS.calledVoice)
			}
		})
	}
}

func TestOrchestrationExecuteTTSAction(t *testing.T) {
	mockTTS := &fakeTTS{}
	cfg := &types.Config{}

	o := Orchestration{TTS: mockTTS, HTTP: httpClient.NewHTTPClient(), Cfg: cfg}

	tests := []struct {
		name          string
		action        types.Action
		expectedText  string
		expectedVoice string
		shouldExecute bool
	}{
		{
			name: "TTS action with voice",
			action: types.Action{
				Type:  "tts",
				Text:  "Hello world",
				Voice: "Microsoft David",
			},
			expectedText:  "Hello world",
			expectedVoice: "Microsoft David",
			shouldExecute: true,
		},
		{
			name: "TTS action without voice",
			action: types.Action{
				Type: "tts",
				Text: "Default voice",
			},
			expectedText:  "Default voice",
			expectedVoice: "",
			shouldExecute: true,
		},
		{
			name: "Non-TTS action",
			action: types.Action{
				Type:  "http",
				Label: "HTTP call",
			},
			expectedText:  "",
			expectedVoice: "",
			shouldExecute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTTS.calledText = ""
			mockTTS.calledVoice = ""
			mockTTS.sayCount = 0

			err := o.ExecuteTTSAction(tt.action)
			if err != nil {
				t.Fatalf("ExecuteTTSAction returned unexpected error: %v", err)
			}

			if tt.shouldExecute {
				if mockTTS.calledText != tt.expectedText {
					t.Errorf("Expected text '%s', got '%s'", tt.expectedText, mockTTS.calledText)
				}

				if mockTTS.calledVoice != tt.expectedVoice {
					t.Errorf("Expected voice '%s', got '%s'", tt.expectedVoice, mockTTS.calledVoice)
				}

				if mockTTS.sayCount != 1 {
					t.Errorf("Expected Say to be called once, got %d", mockTTS.sayCount)
				}
			} else {
				if mockTTS.sayCount != 0 {
					t.Errorf("Expected Say not to be called for non-TTS action, but was called %d times",
						mockTTS.sayCount)
				}
			}
		})
	}
}

func TestOrchestrationExecuteHTTPAction(t *testing.T) {
	requestReceived := false
	lastMethod := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		lastMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockTTS := &fakeTTS{}
	cfg := &types.Config{}
	o := Orchestration{TTS: mockTTS, HTTP: httpClient.NewHTTPClient(), Cfg: cfg}

	tests := []struct {
		name           string
		action         types.Action
		shouldExecute  bool
		expectedMethod string
	}{
		{
			name: "HTTP GET action",
			action: types.Action{
				Type:   "http",
				Method: "GET",
				URL:    server.URL,
			},
			shouldExecute:  true,
			expectedMethod: "GET",
		},
		{
			name: "HTTP POST action",
			action: types.Action{
				Type:   "http",
				Method: "POST",
				URL:    server.URL,
				Body:   `{"test": "data"}`,
			},
			shouldExecute:  true,
			expectedMethod: "POST",
		},
		{
			name: "HTTP default method (POST)",
			action: types.Action{
				Type: "http",
				URL:  server.URL,
				Body: `{"default": "method"}`,
			},
			shouldExecute:  true,
			expectedMethod: "POST",
		},
		{
			name: "Non-HTTP action",
			action: types.Action{
				Type:  "tts",
				Label: "Not HTTP",
			},
			shouldExecute:  false,
			expectedMethod: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestReceived = false
			lastMethod = ""

			err := o.ExecuteHTTPAction(tt.action)
			if err != nil {
				t.Fatalf("ExecuteHTTPAction returned unexpected error: %v", err)
			}

			if tt.shouldExecute && !requestReceived {
				t.Error("Expected HTTP request to be made, but it wasn't")
			}

			if !tt.shouldExecute && requestReceived {
				t.Error("Expected no HTTP request for non-HTTP action, but one was made")
			}

			if tt.shouldExecute && lastMethod != tt.expectedMethod {
				t.Errorf("Expected HTTP method %s, got %s", tt.expectedMethod, lastMethod)
			}
		})
	}
}

func TestOrchestrationMultipleCalls(t *testing.T) {
	mockTTS := &fakeTTS{}
	cfg := &types.Config{}

	o := Orchestration{TTS: mockTTS, HTTP: httpClient.NewHTTPClient(), Cfg: cfg}

	// Appelle Say plusieurs fois
	_ = o.Say("First")
	_ = o.Say("Second")
	_ = o.SayWithVoice("Third", "Voice1")

	if mockTTS.sayCount != 3 {
		t.Errorf("Expected 3 calls, got %d", mockTTS.sayCount)
	}

	// Le dernier appel devrait avoir Voice1
	if mockTTS.calledVoice != "Voice1" {
		t.Errorf("Expected last voice to be 'Voice1', got '%s'", mockTTS.calledVoice)
	}

	if mockTTS.calledText != "Third" {
		t.Errorf("Expected last text to be 'Third', got '%s'", mockTTS.calledText)
	}
}

func TestOrchestrationExecuteIRAction(t *testing.T) {
	mockIR := ir.NewMockIRSender()
	cfg := &types.Config{}
	o := Orchestration{TTS: &fakeTTS{}, HTTP: httpClient.NewHTTPClient(), Cfg: cfg, IR: mockIR}

	action := types.Action{
		Type:       "ir",
		IRDevice:   "tv",
		IRCommand:  "power",
		IRProtocol: "NEC",
		IRCode:     "0x20DF10EF",
		IRRepeat:   2,
	}

	if err := o.ExecuteIRAction(action); err != nil {
		t.Fatalf("ExecuteIRAction returned unexpected error: %v", err)
	}

	lastCmd, err := mockIR.GetLastCommand()
	if err != nil {
		t.Fatalf("GetLastCommand failed: %v", err)
	}

	if lastCmd.Device != "tv" || lastCmd.Command != "power" || lastCmd.Protocol != "NEC" || lastCmd.Code != "0x20DF10EF" || lastCmd.Repeat != 2 {
		t.Errorf("IR command mismatch: %+v", lastCmd)
	}
}

func TestOrchestrationExecuteIRAction_NoIRConfigured(t *testing.T) {
	cfg := &types.Config{}
	o := Orchestration{TTS: &fakeTTS{}, HTTP: httpClient.NewHTTPClient(), Cfg: cfg, IR: nil}

	action := types.Action{
		Type: "ir",
	}

	if err := o.ExecuteIRAction(action); err != nil {
		t.Fatalf("ExecuteIRAction should return nil when IR is not configured: %v", err)
	}
}
