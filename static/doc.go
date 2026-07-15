// Package static provides the HTML templates, stylesheet and images the
// frontend serves.
//
// It has two implementations selected by the 'embed' build tag. With the tag,
// the assets are compiled into the binary via go:embed. Without it, they are
// read from the directory named by StaticPath at startup and held in memory,
// which is what makes editing templates without a rebuild possible.
package static
