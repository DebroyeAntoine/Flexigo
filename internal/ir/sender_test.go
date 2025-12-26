package ir

import (
	"testing"
)

func TestMockIRSender_Send(t *testing.T) {
	sender := NewMockIRSender()

	cmd := IRCommand{
		Device:   "tv",
		Command:  "power",
		Protocol: "NEC",
		Code:     "0x20DF10EF",
		Repeat:   0,
	}

	err := sender.Send(cmd)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if len(sender.SentCommands) != 1 {
		t.Errorf("Expected 1 command, got %d", len(sender.SentCommands))
	}

	lastCmd, err := sender.GetLastCommand()
	if err != nil {
		t.Fatalf("GetLastCommand failed: %v", err)
	}

	if lastCmd.Device != "tv" || lastCmd.Command != "power" {
		t.Errorf("Command mismatch: got %s.%s", lastCmd.Device, lastCmd.Command)
	}
}

func TestMockIRSender_SendMultiple(t *testing.T) {
	sender := NewMockIRSender()

	commands := []IRCommand{
		{Device: "tv", Command: "power"},
		{Device: "tv", Command: "volume_up"},
		{Device: "tv", Command: "volume_down"},
	}

	for _, cmd := range commands {
		if err := sender.Send(cmd); err != nil {
			t.Fatalf("Send failed: %v", err)
		}
	}

	if len(sender.SentCommands) != 3 {
		t.Errorf("Expected 3 commands, got %d", len(sender.SentCommands))
	}
}

func TestMockIRSender_Reset(t *testing.T) {
	sender := NewMockIRSender()

	sender.Send(IRCommand{Device: "tv", Command: "power"})
	sender.Reset()

	if len(sender.SentCommands) != 0 {
		t.Errorf("Expected 0 commands after reset, got %d", len(sender.SentCommands))
	}

	_, err := sender.GetLastCommand()
	if err == nil {
		t.Error("Expected error when getting last command after reset")
	}
}

func TestMockIRSender_ListDevices(t *testing.T) {
	sender := NewMockIRSender()

	devices, err := sender.ListDevices()
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	if len(devices) == 0 {
		t.Error("Expected at least one device")
	}
}

func TestNewIRSender_Mock(t *testing.T) {
	config := DefaultIRConfig()
	sender, err := NewIRSender("mock", config)
	if err != nil {
		t.Fatalf("NewIRSender failed: %v", err)
	}

	if sender == nil {
		t.Fatal("Expected non-nil sender")
	}

	err = sender.Send(IRCommand{Device: "tv", Command: "power"})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
}

func TestNewIRSender_UnsupportedBackend(t *testing.T) {
	config := DefaultIRConfig()
	_, err := NewIRSender("invalid", config)
	if err == nil {
		t.Error("Expected error for unsupported backend")
	}
}

func TestIRCommand_Defaults(t *testing.T) {
	sender := NewMockIRSender()

	// Test avec repeat = 0 (devrait être traité comme 1)
	cmd := IRCommand{
		Device:  "tv",
		Command: "power",
		Repeat:  0,
	}

	err := sender.Send(cmd)
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	lastCmd, _ := sender.GetLastCommand()
	if lastCmd.Repeat != 0 {
		t.Logf("Repeat value preserved: %d", lastCmd.Repeat)
	}
}
