package types

type Action struct {
	Label   string            `yaml:"label"`
	Type    string            `yaml:"type"` // "http", "exec", "tts"
	Method  string            `yaml:"method,omitempty"`
	URL     string            `yaml:"url,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
	Body    string            `yaml:"body,omitempty"`
	Text    string            `yaml:"text,omitempty"`
	Command string            `yaml:"command,omitempty"`
	Args    []string          `yaml:"args,omitempty"`
}

type Config struct {
	Blocks []Action `yaml:"blocks"`
}
