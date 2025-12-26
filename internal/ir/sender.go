package ir

import (
	"fmt"
	"log"
)

// IRCommand représente une commande infrarouge
type IRCommand struct {
	Device   string // Ex: "tv", "ac", "stereo"
	Command  string // Ex: "power", "volume_up", "channel_1"
	Protocol string // Ex: "NEC", "RC5", "Samsung"
	Code     string // Code hex ou raw data
	Repeat   int    // Nombre de répétitions (0 = 1 fois)
}

// IRSender interface pour l'envoi de commandes IR
type IRSender interface {
	Send(cmd IRCommand) error
	SendRaw(protocol string, code string, repeat int) error
	ListDevices() ([]string, error)
	Close() error
}

// NewIRSender crée un nouveau sender IR selon le backend spécifié
func NewIRSender(backend string, config IRConfig) (IRSender, error) {
	switch backend {
	case "serial":
		return NewSerialIRSender(config)
	case "mock":
		return NewMockIRSender(), nil
	default:
		return nil, fmt.Errorf("unsupported IR backend: %s", backend)
	}
}

// IRConfig contient la configuration pour l'IR
type IRConfig struct {
	// Serial configuration
	SerialPort string // Ex: "/dev/ttyUSB0", "COM3", "/dev/cu.usbserial"
	BaudRate   int    // Ex: 9600, 115200

	// Common
	CommandsFile string // Fichier de mapping device/command -> code
	Timeout      int    // Timeout en ms
}

// DefaultIRConfig retourne une configuration par défaut
func DefaultIRConfig() IRConfig {
	return IRConfig{
		SerialPort:   "/dev/ttyUSB0",
		BaudRate:     9600,
		CommandsFile: "ir_commands.yaml",
		Timeout:      5000,
	}
}

// LogCommand log une commande IR (utile pour debug)
func LogCommand(cmd IRCommand) {
	log.Printf("IR Command: device=%s, command=%s, protocol=%s, code=%s, repeat=%d",
		cmd.Device, cmd.Command, cmd.Protocol, cmd.Code, cmd.Repeat)
}
