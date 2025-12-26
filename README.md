# Flexigo 🎯

**Flexigo** is an accessible and configurable user interface application designed to facilitate communication and interaction for people with specific needs. The application features a visual scanning navigation system with customizable buttons that can trigger various actions (text-to-speech, HTTP requests, web browsing, etc.).

## ✨ Key Features

- 🎨 **Customizable graphical interface** with colored button grids
- 🔊 **Text-to-Speech (TTS)** with multiple voice support
- 🌐 **HTTP requests** (GET, POST, PUT, DELETE) to interact with APIs
- 🦊 **Integrated browser mode** for web navigation
- ⌨️ **Configurable virtual keyboard**
- 🎯 **Scanning system** with group/row/item navigation
- 🎨 **Customizable colors** for buttons and highlights
- 🔐 **Environment variable support** for secrets

## 🏗️ Architecture

```
flexigo/
├── internal/
│   ├── browser/          # Browser launch management
│   ├── config/           # YAML configuration loading and validation
│   ├── http/             # HTTP client for API requests
│   ├── orchestration/    # Action orchestration (TTS, HTTP, etc.)
│   ├── tts/              # Text-to-speech (Rust TTS wrapper)
│   ├── types/            # Data types and structures
│   └── ui/               # Graphical interface (Fyne)
├── bin/                  # External binaries (flexigo-tts)
├── assets/
│   └─ config.yaml        # Main configuration
├── .env                  # Environment variables (secrets)
└── main.go               # Entry point
```

## 📋 Prerequisites

- **Go 1.21+**
- **Fyne** (GUI framework)
- **flexigo-tts** (Rust binary for text-to-speech)
- A web browser (Chrome/Firefox) for browser mode

### Installing Dependencies

```bash
# Fyne installation (varies by OS)
# Linux
sudo apt-get install gcc libgl1-mesa-dev xorg-dev

# macOS (with Homebrew)
# No additional dependencies needed

# Windows
# Install TDM-GCC or MinGW-w64
```

## 🚀 Installation

1. **Clone the project**
```bash
git clone https://github.com/DebroyeAntoine/flexigo.git
cd flexigo
```

2. **Install Go dependencies**
```bash
go mod download
```

3. **Compile the TTS binary** (or download from releases)
```bash
# If you have the Rust source code
cd tts-rs
cargo build --release
cp target/release/flexigo-tts ../bin/
```

4. **Create the configuration file**
```bash
cp .env.example .env
# Edit .env with your API keys
```

5. **Build and run**
```bash
go run main.go
# or
go build -o flexigo
./flexigo
```

## ⚙️ Configuration

### `config.yaml` File Structure

```yaml
# Default voice for text-to-speech
voice: "Amélie"

# Default colors (optional)
default_color:
  r: 255
  g: 0
  b: 0
  a: 255

default_highlight_color:
  r: 0
  g: 0
  b: 255
  a: 255

# Default TTS voice (optional)
default_voice: "Microsoft David"

blocks:
  - label: "Main Menu"
    type: container
    timer: 1000              # Scan duration in ms
    grid_width: 4            # Number of columns
    grid_height: 8           # Number of rows
    children:
      # ... actions here
```

### Available Action Types

#### 1. **Container** - Sub-action container
```yaml
- label: "Lights Menu"
  type: container
  grid_width: 3
  grid_height: 3
  timer: 1500
  children:
    - label: "Turn on living room"
      type: http
      # ...
```

#### 2. **TTS** - Text-to-speech
```yaml
- label: "Say hello"
  type: tts
  text: "Hello, how are you?"
  voice: "Microsoft Zira"  # Optional
  width: 2
  height: 1
  position:
    x: 0
    y: 0
```

#### 3. **HTTP** - API requests
```yaml
- label: "Turn on light"
  type: http
  method: POST
  url: "http://homeassistant.local:8123/api/services/light/turn_on"
  headers: # Optional
    Authorization: "Bearer ${HOME_ASSISTANT_TOKEN}"
    Content-Type: "application/json"
  body: '{"entity_id": "light.living_room"}'
```

#### 4. **Browser** - Open a website
```yaml
- label: "🦊 Firefox"
  type: browser
  browser_url: "https://mozilla.org"
  browser_path: "/usr/bin/firefox"  # Optional
  width: 2
  height: 2
  position:
    x: 0
    y: 0
```

#### 5. **Keyboard** - Virtual keyboard
```yaml
- label: "Keyboard"
  type: keyboard
  layout:
    - "ABCDEF"
    - "GHIJKL"
    - "MNOPQR"
    - "STUVWX"
    - "YZ"
```

### Common Properties

```yaml
label: "My Button"           # Displayed text
type: "tts"                  # Action type
width: 2                     # Width in cells
height: 1                    # Height in cells
position:                    # Position in grid
  x: 0
  y: 0
color:                       # Custom color
  r: 255
  g: 100
  b: 50
  a: 255
highlight_color:             # Color during scan
  r: 0
  g: 255
  b: 0
  a: 255
group_membership: 0          # Group ID for scanning
timer: 1000                  # Scan duration (containers)
```

## Usage

### Scanning Navigation

1. **Press ENTER** to start scanning
2. The system first scans **groups** (if configured)
3. Then **rows** of the grid
4. Finally **items** in the selected row
5. Press **ENTER** at each step to validate your choice

### Browser Mode

When you activate a `browser` type button:
- The browser opens in full screen
- A small control window appears with an "Exit" button
- Click "Exit" to close the browser and return to Flexigo

### Virtual Keyboard

1. Activate a `keyboard` type button
2. Use scanning to select letters
3. Special buttons available:
   - **Space**: Add a space
   - **Delete**: Remove the last character
   - **Speak**: Pronounce the entered text
   - **← Back**: Return to previous menu

## 🔐 Environment Variables

Create a `.env` file at the project root:

```bash
# Weather APIs
OPENWEATHER_API_KEY=your_api_key

# Home Assistant
HOME_ASSISTANT_URL=http://homeassistant.local:8123
HOME_ASSISTANT_TOKEN=your_long_token

# Webhooks
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/XXX
DISCORD_WEBHOOK_URL=https://discord.com/api/webhooks/XXX

# GitHub
GITHUB_TOKEN=ghp_your_token

# Other
DUMMY_TOKEN=test123
```

Variables are automatically replaced in URLs, headers, and body of HTTP requests using the `${VARIABLE_NAME}` syntax.

## 📚 Usage Examples

### Example 1: Home Automation with Home Assistant

```yaml
- label: "Home Control"
  type: container
  grid_width: 3
  grid_height: 2
  children:
    - label: "💡 Living Room Light"
      type: http
      method: POST
      url: "${HOME_ASSISTANT_URL}/api/services/light/toggle"
      headers:
        Authorization: "Bearer ${HOME_ASSISTANT_TOKEN}"
      body: '{"entity_id": "light.living_room"}'
      
    - label: "🌡️ Temperature"
      type: http
      method: GET
      url: "${HOME_ASSISTANT_URL}/api/states/sensor.living_room_temperature"
      headers:
        Authorization: "Bearer ${HOME_ASSISTANT_TOKEN}"
```

### Example 2: Quick Communication

```yaml
- label: "Communication"
  type: container
  grid_width: 2
  grid_height: 3
  children:
    - label: "Hello"
      type: tts
      text: "Hello, how are you?"
      
    - label: "Thank you"
      type: tts
      text: "Thank you very much!"
      
    - label: "I need help"
      type: tts
      text: "I need help please"
```

### Example 3: Integration with Public APIs

```yaml
- label: "Information"
  type: container
  grid_width: 3
  grid_height: 2
  children:
    - label: "🌤️ Weather"
      type: http
      method: GET
      url: "https://wttr.in/Paris?format=j1"
      
    - label: "😂 Joke"
      type: http
      method: GET
      url: "https://official-joke-api.appspot.com/random_joke"
      
    - label: "🐱 Cat Fact"
      type: http
      method: GET
      url: "https://catfact.ninja/fact"
```

## 🧪 Tests

```bash
# Run all tests
make test

# Tests with coverage (generates HTML report)
make test-coverage

# Run tests manually
go test ./...
go test -cover ./...

# Tests for a specific package
go test ./internal/http
go test ./internal/config
go test ./internal/orchestration
```

## 🛠️ Development

### Available Make Commands

```bash
make help              # Show all available commands
make build             # Build for current platform
make run               # Build and run
make test              # Run tests
make test-coverage     # Run tests with coverage report
make clean             # Clean build artifacts
make fmt               # Format Go and Rust code
make lint              # Run linters (requires golangci-lint)
make install           # Install system-wide
make release           # Create release archives for all platforms
```

### Module Structure

- **browser**: Browser launch management (Chrome, Firefox)
- **config**: YAML loading + default value application
- **http**: HTTP client with env variable support
- **orchestration**: Coordination between TTS, HTTP and other actions
- **tts**: Interface with Rust binary for text-to-speech
- **types**: Data structure definitions (Action, Config, Color)
- **ui**: Graphical interface with Fyne, scanning management

### Adding a New Action Type

1. Add the type in `types/actions.go`
2. Implement the logic in `orchestration/app.go`
3. Handle execution in `ui/app.go` (`ExecuteAction` function)
4. Add corresponding tests

## 🐛 Troubleshooting

### TTS not working
- Verify that `bin/flexigo-tts` exists and is executable
- Test the binary manually: `./bin/flexigo-tts "Hello"`

### HTTP requests failing
- Check your environment variables in `.env`
- Look at the logs in the console
- Test URLs with curl first

### Browser not launching
- Check the browser path in `browser_path`
- Use default paths by omitting `browser_path`

### Interface not responding
- Make sure you're using the **ENTER** key to validate
- The scan timer might be too fast, increase the value


## 🤝 Contributing

Contributions are welcome! Feel free to:

1. Fork the project
2. Create a branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License. See the `LICENSE` file for details.

## 👤 Author

**Antoine Debroye**
- GitHub: [@DebroyeAntoine](https://github.com/DebroyeAntoine)

## 🙏 Acknowledgments

- [Fyne](https://fyne.io/) - Go GUI framework
- [tts-rs](https://github.com/ndarilek/tts-rs) - Rust TTS library
- All public APIs used in examples

---

**Note**: This project is designed to improve accessibility and autonomy for people with specific communication needs. If you have suggestions for improvements, feel free to open an issue! 💙
