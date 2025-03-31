package main

import (
	"image/color"
	"log"
	"os"

	"github.com/watzon/goshot/background"
	"github.com/watzon/goshot/chrome"
	"github.com/watzon/goshot/content/code"
	"github.com/watzon/goshot/render"
)

func main() {
	// Example showing how to use goshot's redaction feature
	input := `package main

// Example showing various types of sensitive information that
// will be automatically redacted by goshot
func main() {
    // API Keys
    apiKey := "sk_live_51NxXXXXXXXXXXXXXXXXXXXXX"
    
    // AWS Credentials
    awsAccessKey := "AKIAXXXXXXXXXXXXXXXX"
    awsSecretKey := "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    
    // Database Credentials
    dbPassword := "super_secure_password_123!"
    
    // OAuth Token
    githubToken := "ghp_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
    
    // Private Key
    privateKey := ` + "`" + `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC9QFi67tX0hqpx
-----END PRIVATE KEY-----` + "`" + `
}`

	// Create a new canvas with dark background
	canvas := render.NewCanvas().
		WithChrome(chrome.NewMacChrome(
			chrome.MacStyleSequoia,
			chrome.WithTitle("Redaction Example"))).
		WithBackground(
			background.NewColorBackground().
				WithColor(color.RGBA{R: 20, G: 30, B: 40, A: 255}).
				WithPadding(40),
		).
		WithContent(code.DefaultRenderer(input).
			WithLanguage("go").
			WithTheme("dracula").
			WithTabWidth(4).
			WithLineNumbers(true).
			// Enable and configure redaction
			WithRedactionEnabled(true).
			WithRedactionBlurRadius(5.0),
		)

	os.MkdirAll("example_output", 0755)
	err := canvas.SaveAsPNG("example_output/redaction.png")
	if err != nil {
		log.Fatal(err)
	}
}
