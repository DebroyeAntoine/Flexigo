package tts

import "fmt"

type TTSProvider interface {
	Synthesize(text string) ([]byte, error)
}

func NewTTSProvider(provider string) (TTSProvider, error) {
	switch provider {
	case "google":
		return NewGoogleTTS(), nil
	default:
		return nil, fmt.Errorf("provider TTS non supporté: %s", provider)
	}
}
