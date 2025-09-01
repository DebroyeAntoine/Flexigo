package tts

import (
	"bytes"
	"os"
	"os/exec"
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
	// Simule un "stdout"
	_, _ = bytes.NewBufferString("🔊 fake speaking").WriteTo(os.Stdout)
	os.Exit(0)
}

func TestRustTTS_Say(t *testing.T) {
	// On remplace execCommand par notre fake
	execCommand = fakeExecCommand
	defer func() { execCommand = exec.Command }() // restore

	tts := NewRustTTS("fake/path")
	err := tts.Say("Hello world")
	if err != nil {
		t.Fatalf("Say returned unexpected error: %v", err)
	}
}
