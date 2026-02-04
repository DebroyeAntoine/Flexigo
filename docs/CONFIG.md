# Configuration Reference (config.yaml)

This document describes the Flexigo configuration format, its defaults, and all supported action types.

## Overview
The main configuration file lives at:
- `~/.config/Flexigo/config.yaml` (Linux/macOS)
- `%AppData%/Flexigo/config.yaml` (Windows)

If no file exists, a default config is written from the embedded assets.

The root config supports:
- Global defaults (colors, voice, background, IR settings)
- A `blocks` list, which represents the root container (first block) and its children

## Root Fields

```yaml
# Optional global defaults
default_color: { r: 255, g: 0, b: 0, a: 255 }
default_highlight_color: { r: 0, g: 0, b: 255, a: 255 }
default_background: { r: 0, g: 0, b: 0, a: 255 }
default_voice: "Microsoft David"

# Optional legacy field (used by UI voice selection)
voice: "Microsoft David"

# Optional IR settings
ir_backend: "serial"          # "serial" or "mock"
ir_serial_port: "/dev/ttyUSB0" # required for serial
ir_baud_rate: 9600             # default if missing

# Required
blocks:
  - label: "Root"
    type: container
    grid_width: 4
    grid_height: 4
    children: []
```

### Defaults Applied Automatically
- **Timer**: if missing, defaults to `1000ms` (per container and children).
- **Colors**: if missing, defaults are applied recursively to children.
- **Voice**: if `default_voice` is set, all `tts` actions inherit it unless overridden.
- **Group membership**: if missing on container children, defaults to `0`.

## Common Action Fields
All actions can use these base fields:

```yaml
label: "Lights On"
type: "http"  # required
width: 1       # default 1
height: 1      # default 1
position: { x: 0, y: 0 }

# Optional for grouping and scanning
group_membership: 0
# Optional per-action colors
color: { r: 255, g: 0, b: 0, a: 255 }
highlight_color: { r: 0, g: 0, b: 255, a: 255 }
background: { r: 0, g: 0, b: 0, a: 255 }

# Optional scan interval
timer: 1000
```

## Action Types

### 1) Container
A container defines a grid and holds child actions.

```yaml
- label: "Home"
  type: container
  grid_width: 4
  grid_height: 4
  children:
    - label: "Say"
      type: tts
      text: "Hello"
```

Fields:
- `grid_width`, `grid_height` are required for proper layout.
- `children` is required and can be nested recursively.

### 2) TTS (Text-to-Speech)

```yaml
- label: "Speak"
  type: tts
  text: "Hello world"
  voice: "Microsoft Zira" # optional
```

Fields:
- `text` (required)
- `voice` (optional; uses `default_voice` if not set)

### 3) HTTP (Smart Home / APIs)

```yaml
- label: "Lights On"
  type: http
  method: POST
  url: "http://homeassistant.local:8123/api/services/light/turn_on"
  headers:
    Authorization: "Bearer ${HA_TOKEN}"
  body: '{"entity_id": "light.living_room"}'
```

Fields:
- `method` (optional; defaults to `POST`)
- `url` (required)
- `headers` (optional)
- `body` (optional)

Environment variables can be expanded using `${VAR_NAME}` in `url`, `headers`, and `body`.

### 4) IR (Infrared)

```yaml
- label: "TV Power"
  type: ir
  ir_device: "tv"
  ir_command: "power"
  ir_protocol: "NEC"
  ir_code: "0x20DF10EF"
  ir_repeat: 2
```

Fields:
- `ir_device`, `ir_command` (optional labels)
- `ir_protocol` (optional; defaults to `NEC` in serial sender)
- `ir_code` (optional; hex code string)
- `ir_repeat` (optional)

If `ir_backend` is `serial`, `ir_serial_port` must be set.

### 5) Browser

```yaml
- label: "Google"
  type: browser
  browser_url: "https://google.com"
  position: { x: 0, y: 0 }
  width: 2
  height: 1
```

Fields:
- `browser_url` (required)

### 6) Keyboard
A keyboard action opens a virtual keyboard using a layout.

```yaml
- label: "Keyboard"
  type: keyboard
  layout:
    - "EANRCV"
    - "JILPHW"
    - "SUDGK"
```

Special actions are automatically added: `space`, `delete`, `speak`.

### 7) Navigation Helpers
These are created internally by the UI, but can exist in configs:
- `back` (return to previous container)
- `char`, `space`, `delete`, `speak` (keyboard actions)

## Full Example

```yaml
default_color: { r: 255, g: 0, b: 0, a: 255 }
default_highlight_color: { r: 0, g: 0, b: 255, a: 255 }
default_background: { r: 0, g: 0, b: 0, a: 255 }
default_voice: "Microsoft David"

ir_backend: "serial"
ir_serial_port: "/dev/ttyUSB0"

blocks:
  - label: "Root"
    type: container
    grid_width: 4
    grid_height: 4
    children:
      - label: "Say Hello"
        type: tts
        text: "Hello world"
      - label: "Lights On"
        type: http
        method: POST
        url: "http://homeassistant.local:8123/api/services/light/turn_on"
        headers:
          Authorization: "Bearer ${HA_TOKEN}"
        body: '{"entity_id": "light.living_room"}'
      - label: "TV Power"
        type: ir
        ir_protocol: "NEC"
        ir_code: "0x20DF10EF"
      - label: "Browser"
        type: browser
        browser_url: "https://google.com"
      - label: "Keyboard"
        type: keyboard
        layout:
          - "EANRCV"
          - "JILPHW"
          - "SUDGK"
```
