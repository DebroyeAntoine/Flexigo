package types

type Color struct {
	R uint8 `yaml:"r"`
	G uint8 `yaml:"g"`
	B uint8 `yaml:"b"`
	A uint8 `yaml:"a"`
}

type Action struct {
	Label           string            `yaml:"label"`
	Type            string            `yaml:"type"` // "http", "exec", "tts", "container", "keyboard", "char", "speak", "space", "delete", "back"
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
	Color           *Color            `yaml:"color,omitempty"`           // Couleur du bouton
	HighlightColor  *Color            `yaml:"highlight_color,omitempty"` // Couleur lors du scan
}

type Config struct {
	Blocks         []Action `yaml:"blocks"`
	DefaultColor   *Color   `yaml:"default_color,omitempty"`           // Couleur par défaut pour tous les boutons
	HighlightColor *Color   `yaml:"default_highlight_color,omitempty"` // Couleur de highlight par défaut
	DefaultVoice   string   `yaml:"default_voice,omitempty"`           // Voix TTS par défaut
}

type Position struct {
	X int `yaml:"x"`
	Y int `yaml:"y"`
}

func (c *Color) ToImageColor() interface{} {
	type RGBA struct {
		R, G, B, A uint8
	}
	return RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

func DefaultButtonColor() *Color {
	return &Color{R: 255, G: 0, B: 0, A: 255}
}

func DefaultHighlightColor() *Color {
	return &Color{R: 0, G: 0, B: 255, A: 255}
}

