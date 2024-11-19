# 🎨 Goshot

[![Go Reference](https://pkg.go.dev/badge/github.com/watzon/goshot.svg)](https://pkg.go.dev/github.com/watzon/goshot)
[![Go Report Card](https://goreportcard.com/badge/github.com/watzon/goshot)](https://goreportcard.com/report/github.com/watzon/goshot)
[![License](https://img.shields.io/github/license/watzon/goshot)](https://github.com/watzon/goshot/blob/main/LICENSE)

Goshot is a powerful Go library and CLI tool for creating beautiful code screenshots with customizable window chrome, syntax highlighting, and backgrounds. Similar to [Carbon](https://carbon.now.sh) and [Silicon](https://github.com/Aloxaf/Silicon), Goshot allows you to create stunning visual representations of your code snippets for documentation, presentations, or social media sharing.

## ✨ Features

- 🎨 Beautiful syntax highlighting with multiple themes
- 🖼 Customizable window chrome (macOS, Windows, Linux styles)
- 🌈 Various background options (solid colors, gradients, images)
- 🔤 Custom font support
- 📏 Adjustable padding and margins
- 💾 Multiple export formats (PNG, JPEG)
- 🛠 Both CLI and library interfaces

## 📦 Installation

### CLI Tool

```bash
# Install without bundled fonts (uses system fonts)
go install github.com/watzon/goshot/cmd/goshot@latest

# Install with bundled fonts
go install -tags bundle_fonts github.com/watzon/goshot/cmd/goshot@latest
```

### Library

```bash
# Basic installation
go get github.com/watzon/goshot

# When building your application with bundled fonts
go build -tags bundled
```

## 🚀 Usage

### CLI

```bash
# Generate screenshot from a file
goshot code.go -o screenshot.png

# Customize the output
goshot code.go \
  --language go \
  --theme dracula \
  --background "#1a1b26" \
  --window-style mac \
  --padding 32 \
  -o screenshot.png

# Read from stdin
cat code.go | goshot --language go -o screenshot.png
```

### Library

```go
package main

import (
    "github.com/watzon/goshot"
)

func main() {
    shot := goshot.New(&goshot.Config{
        Code: `func main() {
            fmt.Println("Hello, World!")
        }`,
        Language: "go",
        Theme: "dracula",
        Background: goshot.Background{
            Type: goshot.BackgroundSolid,
            Color: "#1a1b26",
        },
        WindowStyle: goshot.WindowStyleMac,
    })

    if err := shot.SaveToPNG("code.png"); err != nil {
        log.Fatal(err)
    }
}
```

## 📁 Project Structure

```
.
├── cmd/
│   └── goshot/          # CLI implementation
├── pkg/
│   ├── window/          # Window styling and rendering
│   │   ├── chrome.go    # Window chrome rendering
│   │   └── style.go     # Window styles (mac, windows, linux)
│   ├── syntax/          # Syntax highlighting
│   │   ├── highlight.go # Code highlighting implementation
│   │   └── theme.go     # Theme definitions and loading
│   ├── background/      # Background processing
│   │   ├── color.go     # Solid color backgrounds
│   │   ├── gradient.go  # Gradient backgrounds
│   │   └── image.go     # Image backgrounds
│   ├── fonts/           # Font loading and management
│   │   ├── fonts.go     # Core font functionality
│   │   ├── fonts_bundled.go   # Bundled font support
│   │   └── fonts_nobundled.go # Fallback for bundled fonts
│   │   └── bundled/     # Bundled font files
│   └── render/          # Final image composition
│       ├── canvas.go    # Main rendering canvas
│       └── export.go    # Export functionality
├── examples/            # Example usage
├── go.mod
├── go.sum
└── README.md
```

## 🗺 Roadmap

### Phase 1: Core Functionality
- ✅ Set up project structure and dependencies
- ✅ Implement basic syntax highlighting using Chroma
- ✅ Add font loading support, including bundled fonts
- [ ] Create basic window chrome rendering
- [ ] Implement solid color backgrounds
- [ ] Add PNG export functionality
- [ ] Create basic CLI interface

### Phase 2: Enhanced Features
- [ ] Add gradient background support
- [ ] Implement image background support
- [ ] Add window style variations (macOS, Windows, Linux)
- [ ] Implement custom font support
- [ ] Add JPEG export functionality
- [ ] Create comprehensive CLI interface

### Phase 3: Polish and Extensions
- [ ] Add more syntax highlighting themes
- [ ] Implement shadow effects
- [ ] Add line number support
- [ ] Create window title customization
- [ ] Add watermark support
- [ ] Implement padding and margin controls

### Phase 4: Documentation and Examples
- [ ] Write comprehensive documentation
- [ ] Create example gallery
- [ ] Add integration tests
- [ ] Create usage examples
- [ ] Add benchmarks

## 📝 License

MIT License - see [LICENSE](LICENSE) for details
