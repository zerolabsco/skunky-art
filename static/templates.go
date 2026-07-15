//go:build embed

package static

import "embed"

// Templates is the asset filesystem compiled into the binary.
//
//go:embed *
var Templates embed.FS

// Enabled reports that assets are embedded in this build.
var Enabled bool = true

// StaticPath is accepted for parity with the non-embed build, where it names the
// directory assets are read from. It is ignored here.
var StaticPath string

// CopyTemplatesToMemory is a no-op in this build: the assets are already embedded.
func CopyTemplatesToMemory() {
	_ = StaticPath
}
