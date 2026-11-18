// Package templates holds the HTML pages, embedded into the binary.
package templates

import "embed"

// FS contains every .html file in this directory.
//
//go:embed *.html
var FS embed.FS
