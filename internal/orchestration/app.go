package orchestration

import (
	"fmt"

	"github.com/DebroyeAntoine/flexigo/internal/tts"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

type Orchestration struct {
	TTS tts.TTSProvider
	// Player audio.AudioPlayer
	Cfg *types.Config
	// éventuellement un queue/lock pour éviter chevauchements
}

func (a *Orchestration) Say(text string) error {
	// 1) a.Player.Stop() pour couper l’éventuelle lecture en cours
	data, err := a.TTS.Synthesize(text)
	if err != nil {
		return err
	}
	fmt.Println("TTS reçu :", len(data), "octets")
	return nil
}
