package orchestration

import (
	"github.com/DebroyeAntoine/flexigo/internal/tts"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

type Orchestration struct {
	TTS tts.TTSProvider
	Cfg *types.Config
}

// Say parle le texte avec la voix par défaut
func (a *Orchestration) Say(text string) error {
	return a.TTS.Say(text)
}

// SayWithVoice parle le texte avec une voix spécifique
func (a *Orchestration) SayWithVoice(text string, voice string) error {
	return a.TTS.SayWithVoice(text, voice)
}

// ExecuteTTSAction exécute une action TTS en utilisant la voix configurée
func (a *Orchestration) ExecuteTTSAction(action types.Action) error {
	if action.Type != "tts" {
		return nil
	}

	if action.Voice != "" {
		return a.SayWithVoice(action.Text, action.Voice)
	}

	return a.Say(action.Text)
}

