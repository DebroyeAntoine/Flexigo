package ir

import (
	"fmt"
	"strings"
	"time"

	"go.bug.st/serial"
)

// SerialIRSender implémentation pour modules IR série (Arduino, ESP32, etc.)
// Compatible tous OS (Windows, Linux, macOS)
type SerialIRSender struct {
	config IRConfig
	port   serial.Port
}

// NewSerialIRSender crée un nouveau sender série
func NewSerialIRSender(config IRConfig) (*SerialIRSender, error) {
	sender := &SerialIRSender{
		config: config,
	}

	if err := sender.connect(); err != nil {
		return nil, fmt.Errorf("failed to open serial port: %w", err)
	}

	return sender, nil
}

// connect ouvre le port série
func (s *SerialIRSender) connect() error {
	mode := &serial.Mode{
		BaudRate: s.config.BaudRate,
		Parity:   serial.NoParity,
		DataBits: 8,
		StopBits: serial.OneStopBit,
	}

	port, err := serial.Open(s.config.SerialPort, mode)
	if err != nil {
		return fmt.Errorf("could not open %s: %w", s.config.SerialPort, err)
	}

	s.port = port

	// Attend que le device soit prêt
	time.Sleep(2 * time.Second)

	return nil
}

// Send envoie une commande IR via le port série
// Format attendu par le module Arduino: sendIR:PROTOCOL,ADDRESS,COMMAND\n
// Exemple: sendIR:NEC,20DF,10EF\n
func (s *SerialIRSender) Send(cmd IRCommand) error {
	if s.port == nil {
		if err := s.connect(); err != nil {
			return err
		}
	}

	LogCommand(cmd)

	protocol := cmd.Protocol
	if protocol == "" {
		protocol = "NEC" // Protocole par défaut
	}

	// Parse le code pour extraire address et command
	// Format attendu du code: "0x20DF10EF" (32 bits NEC)
	// On split en address (16 bits) et command (16 bits)
	var address, command string

	if cmd.Code != "" {
		// Nettoie le préfixe 0x si présent
		code := cmd.Code
		if len(code) > 2 && code[:2] == "0x" {
			code = code[2:]
		}

		// Pour NEC/Samsung: 32 bits = 16 bits addr + 16 bits cmd
		if len(code) >= 4 {
			// Prend les 4 premiers caractères comme address
			address = code[:len(code)/2]
			command = code[len(code)/2:]
		} else {
			// Code trop court, utilise comme command avec address 0
			address = "0"
			command = code
		}
	} else {
		// Fallback: construit depuis device/command
		address = "0"
		command = fmt.Sprintf("%s_%s", cmd.Device, cmd.Command)
	}

	// Format Arduino: sendIR:PROTOCOL,ADDRESS,COMMAND\n
	irCommand := fmt.Sprintf("sendIR:%s,%s,%s\n",
		strings.ToUpper(protocol), address, command)

	_, err := s.port.Write([]byte(irCommand))
	if err != nil {
		s.port = nil // Force reconnection
		return fmt.Errorf("failed to write to serial port: %w", err)
	}

	// Lit la réponse du module
	buf := make([]byte, 128)
	s.port.SetReadTimeout(time.Duration(s.config.Timeout) * time.Millisecond)
	n, err := s.port.Read(buf)
	if err == nil && n > 0 {
		response := strings.TrimSpace(string(buf[:n]))

		// Arduino répond "OK:SEND" en cas de succès
		if strings.HasPrefix(response, "OK:") {
			return nil
		}

		// Erreur du module
		if strings.HasPrefix(response, "ERR:") {
			return fmt.Errorf("IR module error: %s", response)
		}
	}

	// Pas de réponse n'est pas forcément une erreur
	return nil
}

// SendRaw envoie des données IR brutes
// Format: sendIR:RAW,timing1,timing2,timing3,...\n
func (s *SerialIRSender) SendRaw(protocol string, code string, repeat int) error {
	if s.port == nil {
		if err := s.connect(); err != nil {
			return err
		}
	}

	// Pour RAW, le "code" contient les timings séparés par des virgules
	command := fmt.Sprintf("sendIR:RAW,%s\n", code)

	_, err := s.port.Write([]byte(command))
	if err != nil {
		s.port = nil
		return fmt.Errorf("failed to send raw IR: %w", err)
	}

	// Lit la réponse
	buf := make([]byte, 128)
	s.port.SetReadTimeout(time.Duration(s.config.Timeout) * time.Millisecond)
	n, err := s.port.Read(buf)
	if err == nil && n > 0 {
		response := strings.TrimSpace(string(buf[:n]))
		if strings.HasPrefix(response, "OK:") {
			return nil
		}
		if strings.HasPrefix(response, "ERR:") {
			return fmt.Errorf("IR module error: %s", response)
		}
	}

	return nil
}

// ListDevices liste les ports série disponibles
func (s *SerialIRSender) ListDevices() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to list serial ports: %w", err)
	}
	return ports, nil
}

// Close ferme le port série
func (s *SerialIRSender) Close() error {
	if s.port != nil {
		return s.port.Close()
	}
	return nil
}

// ReceiveIR met l'Arduino en mode réception pendant 5 secondes
// et retourne tous les codes IR reçus
func (s *SerialIRSender) ReceiveIR() ([]IRCommand, error) {
	if s.port == nil {
		if err := s.connect(); err != nil {
			return nil, err
		}
	}

	// Envoie la commande recvIR
	_, err := s.port.Write([]byte("recvIR\n"))
	if err != nil {
		return nil, fmt.Errorf("failed to start IR receive mode: %w", err)
	}

	// Lit la confirmation
	buf := make([]byte, 128)
	s.port.SetReadTimeout(1 * time.Second)
	n, err := s.port.Read(buf)
	if err != nil || !strings.Contains(string(buf[:n]), "OK:RECV") {
		return nil, fmt.Errorf("failed to enter receive mode")
	}

	// Lit les codes IR pendant ~5 secondes
	// Format de réponse: IR:PROTOCOL,ADDRESS,COMMAND
	// ou IR:RAW,timing1,timing2,...
	commands := []IRCommand{}
	timeout := time.After(6 * time.Second) // Un peu plus que les 5s de l'Arduino

	for {
		select {
		case <-timeout:
			return commands, nil

		default:
			s.port.SetReadTimeout(500 * time.Millisecond)
			n, err := s.port.Read(buf)
			if err != nil {
				continue
			}

			line := strings.TrimSpace(string(buf[:n]))

			// Ignore keepalive et autres messages
			if line == "KA" || line == "" {
				continue
			}

			// Timeout du module
			if strings.Contains(line, "ERR:TIMEOUT") {
				return commands, nil
			}

			// Parse le code IR reçu
			if strings.HasPrefix(line, "IR:") {
				parts := strings.Split(line[3:], ",")
				if len(parts) >= 1 {
					cmd := IRCommand{
						Protocol: parts[0],
					}

					if parts[0] == "RAW" {
						// Code RAW: stocke tous les timings
						if len(parts) > 1 {
							cmd.Code = strings.Join(parts[1:], ",")
						}
					} else if len(parts) >= 3 {
						// Code standard: PROTOCOL,ADDRESS,COMMAND
						cmd.Code = fmt.Sprintf("0x%s%s", parts[1], parts[2])
					}

					commands = append(commands, cmd)
					fmt.Printf("Received IR: %s", line)
				}
			}
		}
	}
}

func (s *SerialIRSender) ListenForEvents(callback func(string)) {
	if s.port == nil {
		return
	}

	go func() {
		buf := make([]byte, 128)
		var line string
		for {
			n, err := s.port.Read(buf)
			if err != nil {
				// En cas d'erreur, on tente de se reconnecter après un délai
				time.Sleep(2 * time.Second)
				s.connect()
				continue
			}
			if n > 0 {
				line += string(buf[:n])
				if strings.Contains(line, "\n") {
					parts := strings.Split(line, "\n")
					// On traite toutes les lignes complètes sauf la dernière (potentiellement incomplète)
					for i := 0; i < len(parts)-1; i++ {
						msg := strings.TrimSpace(parts[i])
						if msg != "" && msg != "KA" { // On ignore le KeepAlive
							callback(msg)
						}
					}
					line = parts[len(parts)-1]
				}
			}
		}
	}()
}
