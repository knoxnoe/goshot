//go:build !bundle_fonts

package fonts

func init() {
	embeddedFS = nil
}
