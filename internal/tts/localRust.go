package tts

import (
	"fmt"
	"os/exec"
)

type RustTTS struct {
	binPath string
}

var execCommand = exec.Command

func NewRustTTS(path string) TTSProvider {
	return &RustTTS{binPath: path}
}

// Say parle le texte avec la voix par défaut
func (r *RustTTS) Say(text string) error {
	return r.SayWithVoice(text, "")
}

// SayWithVoice parle le texte avec une voix spécifique
func (r *RustTTS) SayWithVoice(text string, voice string) error {
	var cmd *exec.Cmd

	if voice != "" {
		// Mode avec voix spécifique: tts-rs --voice "VoiceName" "text"
		cmd = execCommand(r.binPath, "--voice", voice, text)
	} else {
		// Mode simple: tts-rs "text"
		cmd = execCommand(r.binPath, text)
	}

	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Println(string(output)) // on log ce que le binaire dit
	}
	if err != nil {
		return fmt.Errorf("rust-tts error: %w\nOutput: %s", err, string(output))
	}
	return nil
}

// ListVoices retourne la liste des voix disponibles
func (r *RustTTS) ListVoices() ([]string, error) {
	cmd := execCommand(r.binPath, "--list-voices")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list voices: %w\nOutput: %s", err, string(output))
	}

	// Parse la sortie pour extraire les noms de voix
	// Format attendu: "  VoiceName (language)"
	// TODO: Implémenter le parsing si nécessaire
	fmt.Println(string(output))
	return nil, nil
}

