package orchestration

import (
	"fmt"

	"github.com/DebroyeAntoine/flexigo/internal/player"
	"github.com/DebroyeAntoine/flexigo/internal/tts"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

type Orchestration struct {
	TTS tts.TTSProvider
	Cfg *types.Config

	Player player.AudioPlayer
}

func (a *Orchestration) Say(text string) error {
	data, err := a.TTS.Synthesize(text)
	if err != nil {
		return err
	}
	fmt.Println("TTS reçu :", len(data), "octets")
	return a.Player.PlayMP3(data)
}
