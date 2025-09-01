package tts

import "fmt"

type TTSProvider interface {
	Say(text string) error
}

func NewTTSProvider(provider string) (TTSProvider, error) {
	switch provider {
	case "local":
		return NewRustTTS("bin/flexigo-tts"), nil
	default:
		return nil, fmt.Errorf("provider TTS non supporté: %s", provider)
	}
}
