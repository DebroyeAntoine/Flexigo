package ir

import (
	"fmt"
	"log"
)

// MockIRSender implémentation mock pour les tests
type MockIRSender struct {
	SentCommands []IRCommand
}

// NewMockIRSender crée un nouveau mock sender
func NewMockIRSender() *MockIRSender {
	return &MockIRSender{
		SentCommands: make([]IRCommand, 0),
	}
}

// Send enregistre la commande (mode mock)
func (m *MockIRSender) Send(cmd IRCommand) error {
	LogCommand(cmd)
	m.SentCommands = append(m.SentCommands, cmd)
	log.Printf("[MOCK IR] Would send: %s.%s (protocol=%s, code=%s, repeat=%d)",
		cmd.Device, cmd.Command, cmd.Protocol, cmd.Code, cmd.Repeat)
	return nil
}

// SendRaw enregistre la commande brute
func (m *MockIRSender) SendRaw(protocol string, code string, repeat int) error {
	return m.Send(IRCommand{
		Protocol: protocol,
		Code:     code,
		Repeat:   repeat,
	})
}

// ListDevices retourne des devices fictifs
func (m *MockIRSender) ListDevices() ([]string, error) {
	return []string{"mock_tv", "mock_ac", "mock_stereo"}, nil
}

// Close ne fait rien en mode mock
func (m *MockIRSender) Close() error {
	return nil
}

// Reset réinitialise l'historique des commandes
func (m *MockIRSender) Reset() {
	m.SentCommands = make([]IRCommand, 0)
}

// GetLastCommand retourne la dernière commande envoyée
func (m *MockIRSender) GetLastCommand() (IRCommand, error) {
	if len(m.SentCommands) == 0 {
		return IRCommand{}, fmt.Errorf("no commands sent")
	}
	return m.SentCommands[len(m.SentCommands)-1], nil
}
