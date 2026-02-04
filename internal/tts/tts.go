package tts

import "fmt"
import (
	"os"
	"path/filepath"
	"runtime"
)

type TTSProvider interface {
	Say(text string) error
	SayWithVoice(text string, voice string) error
	ListVoices() ([]string, error)
}

func getTTSPath() string {
	exePath, err := os.Executable()
	if err != nil {
		return "flexigo-tts"
	}
	rootFolder := filepath.Dir(exePath)

	var candidates []string

	switch runtime.GOOS {
	case "darwin":
		if filepath.Base(filepath.Dir(rootFolder)) == "Contents" {
			candidates = append(candidates, filepath.Join(rootFolder, "..", "Resources", "flexigo-tts"))
		}
	case "windows":
		candidates = append(candidates, filepath.Join(rootFolder, "flexigo-tts.exe"))
	default:
		candidates = append(candidates, filepath.Join(rootFolder, "flexigo-tts"))
	}

	if runtime.GOOS == "windows" {
		candidates = append(candidates, "./flexigo-tts.exe")
		candidates = append(candidates, "./bin/flexigo-tts.exe") // optionnel
	} else {
		candidates = append(candidates, "./flexigo-tts")
		candidates = append(candidates, "./bin/flexigo-tts") // optionnel
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	if runtime.GOOS == "windows" {
		return "flexigo-tts.exe"
	}
	return "flexigo-tts"
}

func NewTTSProvider(provider string) (TTSProvider, error) {
	switch provider {
	case "local":
		path := getTTSPath()

		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("binaire TTS introuvable : %s", path)
		}

		return NewRustTTS(path), nil
	default:
		return nil, fmt.Errorf("provider TTS non supporté: %s", provider)
	}
}
