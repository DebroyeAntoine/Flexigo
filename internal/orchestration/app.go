package orchestration

import (
	httpClient "github.com/DebroyeAntoine/flexigo/internal/http"
	"github.com/DebroyeAntoine/flexigo/internal/ir"
	"github.com/DebroyeAntoine/flexigo/internal/tts"
	"github.com/DebroyeAntoine/flexigo/internal/types"
)

// Orchestration wires the action layer to concrete providers.
type Orchestration struct {
	TTS  tts.TTSProvider
	HTTP *httpClient.HTTPClient
	Cfg  *types.Config
	IR   ir.IRSender
}

// Say parle le texte avec la voix par défaut.
func (a *Orchestration) Say(text string) error {
	return a.TTS.Say(text)
}

// SayWithVoice parle le texte avec une voix spécifique.
func (a *Orchestration) SayWithVoice(text string, voice string) error {
	return a.TTS.SayWithVoice(text, voice)
}

// ExecuteTTSAction exécute une action TTS si le type correspond.
func (a *Orchestration) ExecuteTTSAction(action types.Action) error {
	if action.Type != "tts" {
		return nil
	}

	if action.Voice != "" {
		return a.SayWithVoice(action.Text, action.Voice)
	}

	return a.Say(action.Text)
}

// ExecuteHTTPAction exécute une action HTTP si le type correspond.
func (a *Orchestration) ExecuteHTTPAction(action types.Action) error {
	if action.Type != "http" {
		return nil
	}

	method := action.Method
	if method == "" {
		method = "POST"
	}

	return a.HTTP.ExecuteRequest(method, action.URL, action.Headers, action.Body)
}

// ExecuteIRAction exécute une action IR si le type correspond.
func (a *Orchestration) ExecuteIRAction(action types.Action) error {
	if action.Type != "ir" {
		return nil
	}

	if a.IR == nil {
		return nil // IR non configuré
	}

	cmd := ir.IRCommand{
		Device:   action.IRDevice,
		Command:  action.IRCommand,
		Protocol: action.IRProtocol,
		Code:     action.IRCode,
		Repeat:   action.IRRepeat,
	}

	return a.IR.Send(cmd)
}
