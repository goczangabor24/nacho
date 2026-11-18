// Package static holds the stylesheet, embedded into the binary.
package static

import "embed"

// FS contains every CSS file in this directory.
//
//go:embed *.css
var FS embed.FS
