# Flexigo 🎯

**Flexigo** is a high-performance accessibility framework designed for users with motor impairments. It allows full computer and environmental control using a single switch or key. It combines a multi-level scanning interface, ultra-low latency Text-to-Speech (TTS), smart home integration (HTTP/IR), and a dedicated accessible web browser.

## ✨ Key Features

- 🎯 **3-Level Smart Scanning**: Navigate through **Groups > Rows > Items** to optimize selection speed for large grids.
- 🔊 **Native Rust TTS**: Integrated high-speed Text-to-Speech engine written in Rust for immediate feedback.
- 📺 **Infrared (IR) Control**: Control TVs, Air Conditioners, and other appliances via a dedicated Arduino bridge.
- 🦊 **Electron Browser Mode**: A separate, full-screen web instance specifically built for switch-based navigation.
- ⌨️ **Configurable Virtual Keyboard**: Fully customizable layouts via YAML, including word prediction support.
- 🔌 **IoT & API Orchestration**: Native HTTP client (GET/POST/PUT/DELETE) with environment variable injection for Home Assistant, Philips Hue, etc.
- 🖱️ **Hardware Native Integration**: Direct support for physical switches via Arduino (HID Mouse emulation + Serial communication).

## 🏗️ Project Structure

```text
flexigo/
├── internal/
│   ├── ui/               # Fyne GUI, Scanning logic, Grid rendering
│   ├── orchestration/    # Action coordination (TTS, HTTP, IR)
│   ├── ir/               # Serial communication with Arduino
│   ├── tts/              # Rust TTS binary wrapper
│   ├── config/           # YAML loading and ENV injection
│   └── types/            # Shared data structures
├── browser/              # Electron-based browser source code
├── rust/tts-rs/          # Rust TTS engine source code
├── arduino/              # Arduino sketch for Switch & IR hardware
├── bin/                  # Compiled binaries (TTS, Browser, Flexigo)
└── assets/config.yaml    # Main user configuration
```

## 🔌 Hardware Setup (Arduino)

To use a physical switch and Infrared features, you need an **Arduino Leonardo** or **Pro Micro** (ATmega32U4 based for HID support).

1.  **Flash**: Upload the code in `arduino/` to your board using the Arduino IDE.
2.  **Switch**: Connect your accessibility switch to **PIN 2** (Input Pullup).
3.  **IR LED**: Connect an IR emitter LED to **PIN 9**.
4.  **Configuration**: Enable the serial bridge in your `assets/config.yaml`:
    ```yaml
    ir_backend: "serial"
    ir_serial_port: "/dev/ttyACM0" # Linux/macOS or "COM3" on Windows
    ir_baud_rate: 9600
    ```

## 🚀 Installation & Build

### Prerequisites
- **Go 1.21+**
- **Rust (Cargo)** (for the TTS engine)
- **Node.js & NPM** (for the Browser component)
- **C Compiler (gcc)** (required for Fyne GUI)

### Build everything
The project uses a comprehensive `Makefile` to manage the multi-language build process:

```bash
# 1. Install all dependencies (Go, Rust, Node)
make deps

# 2. Build the entire suite (Go app + Rust TTS + Electron Browser)
make build

# 3. Run Flexigo
./bin/flexigo
```

## ⚙️ Configuration Guide (`config.yaml`)

### Grid & Scanning Logic
Flexigo uses a coordinate system `(x, y)`. The scanning follows this hierarchy:
1.  **Groups**: Defined by `group_membership`. Useful for jumping between functional blocks (e.g., jump from "Menu" to "Keyboard").
2.  **Rows**: Vertical scanning within the selected group.
3.  **Items**: Horizontal scanning within the selected row.

### Action Types

#### **1. Infrared (IR)**
Supports NEC, SAMSUNG, and SONY protocols via the Arduino bridge.
```yaml
- label: "TV Power"
  type: ir
  ir_protocol: "SAMSUNG"
  ir_code: "0x0707"
  position: {x: 3, y: 0}
```

#### **2. Browser**
Launches the Electron-based browser to the specified URL.
```yaml
- label: "Google"
  type: browser
  browser_url: "https://google.com"
  position: {x: 1, y: 1}
  width: 2
  height: 2
```

#### **3. HTTP (Smart Home)**
```yaml
- label: "Lights On"
  type: http
  method: POST
  url: "http://homeassistant.local:8123/api/services/light/turn_on"
  headers:
    Authorization: "Bearer ${HA_TOKEN}"
  body: '{"entity_id": "light.living_room"}'
```

## 🔐 Environment Variables
Create a `.env` file at the root to store secrets:
```bash
HA_TOKEN=your_long_lived_access_token
DUMMY_TOKEN=test_123
```
Variables are automatically injected into HTTP URLs, headers, and bodies using the `${VAR_NAME}` syntax.

## ⌨️ Virtual Keyboard
The keyboard layout is fully recursive. You can define custom layouts in YAML:
```yaml
- label: "Keyboard"
  type: keyboard
  layout:
    - "EANRCV"
    - "JILPHW"
    - "SUDGK"
```
Special actions like `space`, `delete`, and `speak` (TTS) are automatically added to the keyboard view.

## 🛠️ Development Commands

| Command | Description |
| :--- | :--- |
| `make build-rust` | Recompile only the Rust TTS engine |
| `make build-browser` | Recompile only the Electron Browser |
| `make build-go` | Recompile only the Go core |
| `make clean` | Remove all binaries and build artifacts |

---

**Author:** Antoine Debroye  
