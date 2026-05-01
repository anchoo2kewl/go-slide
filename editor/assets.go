// Package editor exposes the embedded editor bundle (CSS + JS + templates)
// used by hosts that want to render the go-slide editor inside an admin
// page.
//
// The Day-1 scaffold ships an empty bundle; Day-2 work adds the TipTap
// design-system schema, the WYSIWYG canvas, and the preview iframe wiring.
package editor

import (
	"embed"
	"io/fs"
)

//go:embed css/*.css js/*.js templates/*.gohtml
var embeddedFS embed.FS

// Assets returns the embedded filesystem containing CSS, JS, and templates.
// Use with http.FileServer or to read individual files.
func Assets() fs.FS { return embeddedFS }
