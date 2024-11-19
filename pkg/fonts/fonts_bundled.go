//go:build bundle_fonts

package fonts

import "embed"

//go:embed bundled/*
var bundledFS embed.FS

func init() {
	embeddedFS = &bundledFS
}
