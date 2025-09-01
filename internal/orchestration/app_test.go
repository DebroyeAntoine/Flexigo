package orchestration

import (
	"testing"

	"github.com/DebroyeAntoine/flexigo/internal/types"
)

type fakeTTS struct {
	calledText string
}

func (f *fakeTTS) Say(text string) error {
	f.calledText = text
	return nil
}

func TestOrchestrationSay(t *testing.T) {
	mockTTS := &fakeTTS{}
	cfg := &types.Config{}

	o := Orchestration{TTS: mockTTS, Cfg: cfg}

	err := o.Say("Hello test")
	if err != nil {
		t.Fatalf("Say returned unexpected error: %v", err)
	}

	if mockTTS.calledText != "Hello test" {
		t.Errorf("Expected TTS.Say to be called with 'Hello test', got '%s'", mockTTS.calledText)
	}
}
