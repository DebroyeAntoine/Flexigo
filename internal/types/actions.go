package types

type Action struct {
	Label    string            `yaml:"label"`
	Type     string            `yaml:"type"` // "http", "exec", "tts", "container"
	Method   string            `yaml:"method,omitempty"`
	URL      string            `yaml:"url,omitempty"`
	Headers  map[string]string `yaml:"headers,omitempty"`
	Body     string            `yaml:"body,omitempty"`
	Text     string            `yaml:"text,omitempty"`
	Command  string            `yaml:"command,omitempty"`
	Args     []string          `yaml:"args,omitempty"`
	Children []Action          `yaml:"children,omitempty"` // Sous-blocs récursifs
	Timer    int               `yaml:"timer,omitempty"`
	Layout   []string          `yaml:"layout,omitempty"`
	Width    int               `yaml:"width,omitempty"`
	Height   int               `yaml:"height,omitempty"`
	Position Position          `yaml:"position,omitempty"`
}

type Config struct {
	Blocks []Action `yaml:"blocks"`
}

type Position struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}
