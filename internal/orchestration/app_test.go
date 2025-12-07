package orchestration

import (
	"net/http"
	"net/http/httptest"
	"testing"

	httpClient "github.com/DebroyeAntoine/flexigo/internal/http"
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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestReceived = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockTTS := &fakeTTS{}
	cfg := &types.Config{}
	o := Orchestration{TTS: mockTTS, HTTP: httpClient.NewHTTPClient(), Cfg: cfg}

	tests := []struct {
		name          string
		action        types.Action
		shouldExecute bool
	}{
		{
			name: "HTTP GET action",
			action: types.Action{
				Type:   "http",
				Method: "GET",
				URL:    server.URL,
			},
			shouldExecute: true,
		},
		{
			name: "HTTP POST action",
			action: types.Action{
				Type:   "http",
				Method: "POST",
				URL:    server.URL,
				Body:   `{"test": "data"}`,
			},
			shouldExecute: true,
		},
		{
			name: "Non-HTTP action",
			action: types.Action{
				Type:  "tts",
				Label: "Not HTTP",
			},
			shouldExecute: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestReceived = false

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
