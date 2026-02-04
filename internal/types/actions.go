package types

// Color is a simple RGBA color for UI elements.
type Color struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
	A uint8 `yaml:"a"`
}

// Action describes a UI block and its behavior.
type Action struct {
	Label           string            `yaml:"label"`
	Type            string            `yaml:"type"` // "http", "exec", "tts", "container", "keyboard", "char", "speak", "space", "delete", "back", "ir"
	Method          string            `yaml:"method,omitempty"`
	URL             string            `yaml:"url,omitempty"`
	Headers         map[string]string `yaml:"headers,omitempty"`
	Body            string            `yaml:"body,omitempty"`
	Text            string            `yaml:"text,omitempty"`
	Voice           string            `yaml:"voice,omitempty"` // Voix TTS à utiliser
	Command         string            `yaml:"command,omitempty"`
	Args            []string          `yaml:"args,omitempty"`
	Children        []Action          `yaml:"children,omitempty"` // Sous-blocs récursifs
	Timer           int               `yaml:"timer,omitempty"`    // Temps de scan en ms
	Layout          []string          `yaml:"layout,omitempty"`   // Layout pour clavier
	Width           int               `yaml:"width,omitempty"`
	Height          int               `yaml:"height,omitempty"`
	Position        Position          `yaml:"position,omitempty"`
	GridWidth       int               `yaml:"grid_width,omitempty"`
	GridHeight      int               `yaml:"grid_height,omitempty"`
	GroupMembership *int              `yaml:"group_membership,omitempty"`
	Color           *Color            `yaml:"color,omitempty"`        // Couleur du bouton
	BrowserPath     string            `yaml:"browser_path,omitempty"` // Chemin vers l'exécutable du navigateur
	BrowserURL      string            `yaml:"browser_url,omitempty"`
	HighlightColor  *Color            `yaml:"highlight_color,omitempty"` // Couleur lors du scan
	Background      *Color            `yaml:"background,omitempty"`

	// Infrared fields
	IRDevice   string `yaml:"ir_device,omitempty"`   // Ex: "tv", "ac"
	IRCommand  string `yaml:"ir_command,omitempty"`  // Ex: "power", "volume_up"
	IRProtocol string `yaml:"ir_protocol,omitempty"` // Ex: "NEC", "RC5"
	IRCode     string `yaml:"ir_code,omitempty"`     // Code hex
	IRRepeat   int    `yaml:"ir_repeat,omitempty"`   // Nombre de répétitions
}

// Config is the main YAML configuration for Flexigo.
type Config struct {
	Voice          string   `yaml:"voice,omitempty"` // Voix TTS à utiliser
	Blocks         []Action `yaml:"blocks"`
	DefaultColor   *Color   `yaml:"default_color,omitempty"`           // Couleur par défaut pour tous les boutons
	HighlightColor *Color   `yaml:"default_highlight_color,omitempty"` // Couleur de highlight par défaut
	DefaultVoice   string   `yaml:"default_voice,omitempty"`           // Voix TTS par défaut
	Background     *Color   `yaml:"default_background,omitempty"`      // Couleur de highlight par défaut

	// IR configuration
	IRBackend    string `yaml:"ir_backend,omitempty"`     // "serial", "mock"
	IRSerialPort string `yaml:"ir_serial_port,omitempty"` // Ex: "/dev/ttyUSB0"
	IRBaudRate   int    `yaml:"ir_baud_rate,omitempty"`   // Ex: 9600
	IRLIRCSocket string `yaml:"ir_lirc_socket,omitempty"` // Ex: "/var/run/lirc/lircd"
}

// Position is a grid coordinate.
type Position struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
}

// ToImageColor returns a struct compatible with image color usage.
func (c *Color) ToImageColor() interface{} {
	type RGBA struct {
		R, G, B, A uint8
	}
	return RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// DefaultButtonColor returns the default button color.
func DefaultButtonColor() *Color {
	return &Color{R: 255, G: 0, B: 0, A: 255}
}

// DefaultHighlightColor returns the default highlight color.
func DefaultHighlightColor() *Color {
	return &Color{R: 0, G: 0, B: 255, A: 255}
}

// DefaultBackgroundColor returns the default background color.
func DefaultBackgroundColor() *Color {
	return &Color{R: 0, G: 0, B: 0, A: 255}
}
