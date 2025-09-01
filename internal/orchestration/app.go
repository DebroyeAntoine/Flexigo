package orchestration

import (
	"github.com/DebroyeAntoine/flexigo/internal/tts"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

type Orchestration struct {
	TTS tts.TTSProvider
	Cfg *types.Config
}

func (a *Orchestration) Say(text string) error {
	err := a.TTS.Say(text)
	if err != nil {
		return err
	}
	return nil
}
