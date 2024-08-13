//go:build embed
// +build embed

package static

import "embed"

//go:embed *
var Templates embed.FS
var Enabled bool = true

var StaticPath string

func CopyTemplatesToMemory() {
	_ = StaticPath
}
