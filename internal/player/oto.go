package player

import (
	"bytes"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
)

type OtoPlayer struct {
	ctx *oto.Context
	p   oto.Player
}

func NewOtoPlayer() *OtoPlayer {
	return &OtoPlayer{}
}

func (o *OtoPlayer) PlayMP3(data []byte) error {
	decoder, err := mp3.NewDecoder(bytes.NewReader(data))
	if err != nil {
		return err
	}

	if o.ctx == nil {
		c, ready, err := oto.NewContext(decoder.SampleRate(), 2, 2)
		if err != nil {
			return err
		}
		<-ready
		o.ctx = c
	}

	if o.p != nil {
		o.p.Close()
	}

	o.p = o.ctx.NewPlayer(decoder)
	o.p.Play()

	go func() {
		for o.p.IsPlaying() {
			time.Sleep(50 * time.Millisecond)
		}
		o.p.Close()
	}()

	return nil
}

func (o *OtoPlayer) Stop() error {
	if o.p != nil {
		o.p.Close()
	}
	return nil
}
