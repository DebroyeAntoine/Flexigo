package tts

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// Fake exec.Command pour capturer les args
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

// Test helper process (simule un vrai binaire)
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// Récupère les arguments passés
	args := os.Args
	for i, arg := range args {
		if arg == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) >= 1 && args[0] == "fail" {
		_, _ = bytes.NewBufferString("forced failure").WriteTo(os.Stderr)
		os.Exit(1)
	}

	// Simule différents comportements selon les arguments
	if len(args) >= 2 {
		if args[1] == "--list-voices" {
			_, _ = bytes.NewBufferString("🎵 Available voices:\n  Microsoft David (en-US)\n  Microsoft Zira (en-US)").WriteTo(os.Stdout)
			os.Exit(0)
		}

		if args[1] == "--voice" && len(args) >= 4 {
			// Mode: --voice VoiceName "text"
			voice := args[2]
			text := args[3]
			_, _ = bytes.NewBufferString("🎭 Voice: " + voice + "\n📊 Speaking: " + text).WriteTo(os.Stdout)
			os.Exit(0)
		}

		// Mode simple: "text"
		text := args[1]
		_, _ = bytes.NewBufferString("📊 Speaking: " + text).WriteTo(os.Stdout)
		os.Exit(0)
	}

	os.Exit(0)
}

func TestRustTTS_Say(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	tts := NewRustTTS("fake/path")
	err := tts.Say("Hello world")
	if err != nil {
		t.Fatalf("Say returned unexpected error: %v", err)
	}
}

func TestRustTTS_SayWithVoice(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	tts := NewRustTTS("fake/path")

	tests := []struct {
		name  string
		text  string
		voice string
	}{
		{
			name:  "With custom voice",
			text:  "Hello",
			voice: "Microsoft Zira",
		},
		{
			name:  "With empty voice (default)",
			text:  "World",
			voice: "",
		},
		{
			name:  "With different voice",
			text:  "Test",
			voice: "Microsoft David",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tts.SayWithVoice(tt.text, tt.voice)
			if err != nil {
				t.Fatalf("SayWithVoice returned unexpected error: %v", err)
			}
		})
	}
}

func TestRustTTS_SayWithVoice_Integration(t *testing.T) {
	// Ce test simule l'utilisation réelle
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	tts := NewRustTTS("fake/path")

	// Test 1: Utiliser Say (voix par défaut)
	err := tts.Say("Default voice")
	if err != nil {
		t.Errorf("Say failed: %v", err)
	}

	// Test 2: Utiliser SayWithVoice avec une voix
	err = tts.SayWithVoice("Custom voice", "Microsoft Zira")
	if err != nil {
		t.Errorf("SayWithVoice with voice failed: %v", err)
	}

	// Test 3: Utiliser SayWithVoice sans voix (devrait être comme Say)
	err = tts.SayWithVoice("No voice specified", "")
	if err != nil {
		t.Errorf("SayWithVoice without voice failed: %v", err)
	}
}

func TestRustTTS_ListVoices(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	tts := NewRustTTS("fake/path")

	// Note: ListVoices retourne actuellement nil, nil car le parsing n'est pas implémenté
	// Ce test vérifie juste qu'il n'y a pas d'erreur
	voices, err := tts.ListVoices()
	if err != nil {
		t.Fatalf("ListVoices returned error: %v", err)
	}
	if voices != nil {
		t.Errorf("expected voices to be nil until parsing is implemented, got %v", voices)
	}

	// Quand le parsing sera implémenté, on pourra vérifier:
	_ = voices
	// if len(voices) != 2 {
	//     t.Errorf("expected 2 voices, got %d", len(voices))
	// }
}

func TestRustTTS_ListVoices_Error(t *testing.T) {
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }()

	tts := NewRustTTS("fail")
	voices, err := tts.ListVoices()
	if err == nil {
		t.Fatal("expected error when list voices fails, got nil")
	}
	if voices != nil {
		t.Errorf("expected voices to be nil on error, got %v", voices)
	}
}

func TestRustTTS_ErrorHandling(t *testing.T) {
	// Utilise vraiment exec.Command pour tester les erreurs
	tts := NewRustTTS("/path/that/does/not/exist")

	err := tts.Say("This should fail")
	if err == nil {
		t.Error("expected error when binary doesn't exist, got nil")
	}

	if !strings.Contains(err.Error(), "rust-tts error") {
		t.Errorf("expected error to contain 'rust-tts error', got: %v", err)
	}
}

func TestRustTTS_VoiceErrorHandling(t *testing.T) {
	tts := NewRustTTS("/path/that/does/not/exist")

	err := tts.SayWithVoice("This should fail", "NonExistentVoice")
	if err == nil {
		t.Error("expected error when binary doesn't exist, got nil")
	}
}
