package tts

import (
	"fmt"
	"os/exec"
)

type RustTTS struct {
	binPath string
}

func NewRustTTS(path string) TTSProvider {
	return &RustTTS{binPath: path}
}

func (r *RustTTS) Say(text string) error {
	cmd := exec.Command(r.binPath, text)
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(string(output)) // on log ce que le binaire dit
	}
	if err != nil {
		return fmt.Errorf("rust-tts error: %w\nOutput: %s", err, string(output))
	}
	return nil
}
