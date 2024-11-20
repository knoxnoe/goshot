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

> [!NOTE]  
> This is a work in progress and will be coming very soon.

```bash
# Generate screenshot from a file
goshot code.go -o screenshot.png

# Customize the output
goshot code.go \
  --language go

# Read from stdin
cat code.go | goshot --language go -o screenshot.png
```

### Library

```go
package main

import (
    "image/color"
    "log"

    "github.com/watzon/goshot/pkg/background"
    "github.com/watzon/goshot/pkg/chrome"
    "github.com/watzon/goshot/pkg/render"
)

func main() {
    canvas := render.NewCanvas().
        SetChrome(chrome.NewMacChrome(chrome.WithTitle("Hello World"))).
        SetBackground(
            background.NewGradientBackground(
                background.LinearGradient,
                background.GradientStop{Color: color.RGBA{R: 26, G: 27, B: 38, A: 255}, Position: 0},
                background.GradientStop{Color: color.RGBA{R: 40, G: 42, B: 54, A: 255}, Position: 1},
            ).SetAngle(45).SetPadding(40),
        ).
        SetCodeStyle(&render.CodeStyle{
            Language:        "go",
            Theme:          "dracula",
            TabWidth:       4,
            ShowLineNumbers: true,
        })

    code := `func main() {
        fmt.Println("Hello, World!")
    }`

    if err := canvas.RenderToFile(code, "code.png"); err != nil {
        log.Fatal(err)
    }
}
```

## 🎨 Background Options

Goshot supports various background types to make your code screenshots stand out:

### Solid Color Background

```go
background.NewColorBackground().
    SetColor(color.RGBA{R: 30, G: 30, B: 30, A: 255}).
    SetPadding(40)
```

### Gradient Backgrounds

#### Linear Gradient
```go
background.NewGradientBackground(
    background.LinearGradient,
    background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
    background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
).SetAngle(45).SetPadding(40)
```

#### Radial Gradient
```go
background.NewGradientBackground(
    background.RadialGradient,
    background.GradientStop{Color: color.RGBA{R: 30, G: 30, B: 30, A: 255}, Position: 0},
    background.GradientStop{Color: color.RGBA{R: 60, G: 60, B: 60, A: 255}, Position: 1},
).SetCenter(0.5, 0.5).SetPadding(40)
```

#### Angular Gradient
```go
background.NewGradientBackground(
    background.AngularGradient,
    background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Position: 0},
    background.GradientStop{Color: color.RGBA{R: 0, G: 255, B: 0, A: 255}, Position: 0.33},
    background.GradientStop{Color: color.RGBA{R: 0, G: 0, B: 255, A: 255}, Position: 0.66},
).SetAngle(45).SetPadding(40)
```

#### Diamond Gradient
```go
background.NewGradientBackground(
    background.DiamondGradient,
    background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 255, A: 255}, Position: 0},
    background.GradientStop{Color: color.RGBA{R: 128, G: 0, B: 128, A: 255}, Position: 0.5},
    background.GradientStop{Color: color.RGBA{R: 255, G: 255, B: 0, A: 255}, Position: 1},
).SetCenter(0.5, 0.5).SetPadding(40)
```

#### Spiral Gradient
```go
background.NewGradientBackground(
    background.SpiralGradient,
    background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Position: 0},
    background.GradientStop{Color: color.RGBA{R: 0, G: 0, B: 255, A: 255}, Position: 1},
).SetIntensity(3.0).SetAngle(0).SetPadding(40)
```

#### Square Gradient
```go
background.NewGradientBackground(
    background.SquareGradient,
    background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Position: 0},
    background.GradientStop{Color: color.RGBA{R: 0, G: 0, B: 255, A: 255}, Position: 1},
).SetCenter(0.5, 0.5).SetPadding(40)
```

#### Star Gradient
```go
background.NewGradientBackground(
    background.StarGradient,
    background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 0, A: 255}, Position: 0},
    background.GradientStop{Color: color.RGBA{R: 0, G: 0, B: 255, A: 255}, Position: 1},
).SetIntensity(7).SetAngle(45).SetPadding(40) // 7 points in the star
```

### Image Background

```go
// Load an image
file, _ := os.Open("background.jpg")
img, _, _ := image.Decode(file)

background.NewImageBackground(img).
    SetScaleMode(background.ImageScaleFill).
    SetBlurRadius(3.0).
    SetOpacity(0.9).
    SetPadding(40).
    SetCornerRadius(10)
```

All background types support:
- Padding control
- Corner radius for rounded corners
- Integration with window chrome

Additional features per type:
- **Gradients**: Angle, center point, and intensity control (where applicable)
- **Images**: Scale modes (fit, fill, stretch, tile, and cover), blur effects, and opacity

### Example with Chrome and Code Style

Here's a complete example that combines background, chrome, and code styling:

```go
render.NewCanvas().
    SetChrome(chrome.NewWindows11Chrome(chrome.WithTitle("My App"))).
    SetBackground(
        background.NewGradientBackground(
            background.DiamondGradient,
            background.GradientStop{Color: color.RGBA{R: 255, G: 0, B: 255, A: 255}, Position: 0},
            background.GradientStop{Color: color.RGBA{R: 128, G: 0, B: 128, A: 255}, Position: 0.5},
            background.GradientStop{Color: color.RGBA{R: 255, G: 255, B: 0, A: 255}, Position: 1},
        ).SetCenter(0.5, 0.5).SetPadding(100),
    ).
    SetCodeStyle(&render.CodeStyle{
        Language:            "go",
        Theme:              "dracula",
        TabWidth:           4,
        ShowLineNumbers:    true,
        LineHighlightRanges: []render.LineRange{{Start: 18, End: 26}},
    })
```

## 📁 Project Structure

```
.
├── cmd/
│   ├── goshot/          # CLI implementation
│   └── examples/        # Example code
├── pkg/
│   ├── background/      # Background processing
│   │   ├── background.go # Main background interface
│   │   ├── color.go     # Solid color backgrounds
│   │   ├── gradient.go  # Gradient backgrounds
│   │   └── image.go     # Image backgrounds
│   ├── chrome/          # Window styling and rendering
│   │   ├── chrome.go    # Window chrome rendering
│   │   └── macos.go     # macOS-specific window chrome
│   │   └── windows11.go # Windows 11-specific window chrome
│   │   └── utils.go     # Utility functions
│   ├── fonts/           # Font loading and management
│   │   ├── fonts.go     # Core font functionality
│   │   ├── fonts_bundled.go   # Bundled font support
│   │   └── fonts_nobundled.go # Fallback for bundled fonts
│   │   └── bundled/     # Bundled font files
│   ├── render/          # Final image composition
│   │   ├── canvas.go    # Main rendering canvas
│   │   └── export.go    # Export functionality
│   └── syntax/          # Syntax highlighting
│       ├── syntax.go    # Main syntax interface
│       └── render.go    # Syntax rendering
├── go.mod
├── go.sum
└── README.md
```

## 🗺 Roadmap

### Core Functionality
- ✅ Set up project structure and dependencies
- ✅ Implement basic syntax highlighting using Chroma
- ✅ Add font loading support, including bundled fonts
- ✅ Create basic window chrome rendering
- ✅ Implement solid color backgrounds
- ✅ Add PNG export functionality
- [ ] Create basic CLI interface

### Enhanced Features
- ✅ Add gradient background support
- ✅ Implement image background support
- ✅ Add window style variations (macOS, Windows, Linux)
- ✅ Implement custom font support
- ✅ Add JPEG export functionality
- [ ] Create comprehensive CLI interface

### Polish and Extensions
- [ ] Add support for emojis
- [ ] Implement shadow effects
- ✅ Add line number support
- ✅ Create window title customization
- [ ] Add watermark support
- ✅ Implement padding and margin controls

### Documentation and Examples
- [ ] Write comprehensive documentation
- [ ] Create example gallery
- [ ] Add integration tests
- [ ] Create usage examples
- [ ] Add benchmarks

## 📝 License

MIT License - see [LICENSE](LICENSE) for details
