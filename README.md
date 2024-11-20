# Goshot

<div align="center">
    <img src=".github/example.png">
</div>

<div align="center">
    <a href="https://pkg.go.dev/github.com/watzon/goshot"><img src="https://pkg.go.dev/badge/github.com/watzon/goshot.svg" alt="Go Reference"></a>
    <a href="https://goreportcard.com/report/github.com/watzon/goshot"><img src="https://goreportcard.com/badge/github.com/watzon/goshot" alt="Go Report Card"></a>
    <a href="LICENSE"><img src="https://img.shields.io/github/license/watzon/goshot" alt="License"></a>
</div>

Goshot is a powerful Go library and CLI tool for creating beautiful code screenshots with customizable window chrome, syntax highlighting, and backgrounds. Similar to [Carbon](https://carbon.now.sh) and [Silicon](https://github.com/Aloxaf/Silicon), Goshot allows you to create stunning visual representations of your code snippets for documentation, presentations, or social media sharing.

## ✨ Features

- 🎨 Beautiful syntax highlighting with multiple themes
- 🖼 Customizable window chrome (macOS, Windows, Linux styles)
- 🌈 Various background options (solid colors, gradients, images)
- 🔤 Custom font support
- 📏 Adjustable padding and margins
- 💾 Multiple export formats (PNG, JPEG)
- 🛠 Both CLI and library interfaces

## Quick Start

### Installation

```bash
# Install the CLI tool
go install github.com/watzon/goshot/cmd/goshot@latest

# Install the library
go get github.com/watzon/goshot
```

### Basic Usage

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
    // Create a new canvas with macOS chrome and gradient background
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

    // Render code to file
    code := `func main() {
        fmt.Println("Hello, World!")
    }`
    if err := canvas.RenderToFile(code, "code.png"); err != nil {
        log.Fatal(err)
    }
}
```

## Documentation

For detailed documentation, examples, and guides, please visit our [Wiki](https://github.com/watzon/goshot/wiki):

- [Installation Guide](https://github.com/watzon/goshot/wiki/Installation) - Detailed installation instructions
- [Library Usage](https://github.com/watzon/goshot/wiki/Library-Usage) - Library documentation and examples
- [Configuration](https://github.com/watzon/goshot/wiki/Configuration) - Configuration options and customization
- [Contributing](https://github.com/watzon/goshot/wiki/Contributing) - Guidelines for contributing

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
