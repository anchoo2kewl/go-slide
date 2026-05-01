// Package editor renders the embeddable go-slide editor as an HTML
// fragment that hosts can drop into any admin page.
//
// The fragment includes:
//   - a <link> to the editor chrome CSS (toolbar, sidebar, canvas)
//   - a <style> tag with the selected theme's CSS so the WYSIWYG canvas
//     looks pixel-identical to the rendered presentation page
//   - the import map + <script type=module> bootstrapping TipTap and
//     wiring up GoSlideEditor
//   - a hidden <input> the host's form submits as the slide content
//
// The CSS and JS files referenced via {{.AssetBase}} must be served from
// editor.Assets() (mounted via handler.AssetHandler on the host).
package editor

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/anchoo2kewl/go-slide/theme"
)

//go:embed css/*.css js/*.mjs templates/*.gohtml
var embeddedFS embed.FS

// Assets returns the embedded filesystem containing the editor's CSS, JS
// modules, and templates. Mount with handler.AssetHandler.
func Assets() embed.FS { return embeddedFS }

// Config controls what the rendered editor includes.
type Config struct {
	// AssetBase is the URL prefix where the editor's assets are served.
	// Required (e.g. "/slides/assets").
	AssetBase string

	// PreviewEndpoint is the POST URL the editor calls for live preview.
	// Required (e.g. "/slides/preview").
	PreviewEndpoint string

	// AutosaveEndpoint is the POST URL the editor calls to autosave.
	// Optional; if empty, only manual Save fires.
	AutosaveEndpoint string

	// Theme is the design system the editor canvas uses for styling.
	// Required.
	Theme theme.Theme

	// InitialContent is the existing slide HTML loaded on mount.
	// May be empty for a new deck.
	InitialContent string

	// OutputFieldID is the DOM ID of the hidden <input> the host's form
	// submits as the slide content. Default: "go-slide-content".
	OutputFieldID string
	// OutputFieldName is the form field name. Default: "content".
	OutputFieldName string
	// FormID is the host form's DOM ID. Default: "slide-form".
	FormID string
}

// templateData is the data passed to the editor template.
type templateData struct {
	AssetBase          string
	Theme              theme.Theme
	ThemeCSS           template.CSS
	InitialContentJS   template.JS
	ThemeIDJS          template.JS
	PreviewEndpointJS  template.JS
	AssetBaseJS        template.JS
	AutosaveEndpointJS template.JS
	OutputFieldIDJS    template.JS
	FormIDJS           template.JS
	OutputFieldID      string
	OutputFieldName    string
}

// Render returns the editor's HTML fragment.
func Render(cfg Config) (template.HTML, error) {
	if cfg.AssetBase == "" {
		return "", fmt.Errorf("editor: AssetBase is required")
	}
	if cfg.PreviewEndpoint == "" {
		return "", fmt.Errorf("editor: PreviewEndpoint is required")
	}
	if cfg.Theme.ID == "" {
		return "", fmt.Errorf("editor: Theme is required")
	}
	if cfg.OutputFieldID == "" {
		cfg.OutputFieldID = "go-slide-content"
	}
	if cfg.OutputFieldName == "" {
		cfg.OutputFieldName = "content"
	}
	if cfg.FormID == "" {
		cfg.FormID = "slide-form"
	}

	tpl, err := template.ParseFS(embeddedFS, "templates/*.gohtml")
	if err != nil {
		return "", fmt.Errorf("editor: parse template: %w", err)
	}

	data := templateData{
		AssetBase:          cfg.AssetBase,
		Theme:              cfg.Theme,
		ThemeCSS:           cfg.Theme.RootCSS(),
		InitialContentJS:   jsString(cfg.InitialContent),
		ThemeIDJS:          jsString(cfg.Theme.ID),
		PreviewEndpointJS:  jsString(cfg.PreviewEndpoint),
		AssetBaseJS:        jsString(cfg.AssetBase),
		AutosaveEndpointJS: jsString(cfg.AutosaveEndpoint),
		OutputFieldIDJS:    jsString(cfg.OutputFieldID),
		FormIDJS:           jsString(cfg.FormID),
		OutputFieldID:      cfg.OutputFieldID,
		OutputFieldName:    cfg.OutputFieldName,
	}

	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "editor.gohtml", data); err != nil {
		return "", fmt.Errorf("editor: execute template: %w", err)
	}
	return template.HTML(buf.String()), nil
}

// jsString JSON-encodes a string so it's safe to drop into a JS literal.
// Returns template.JS so html/template won't double-escape it.
func jsString(s string) template.JS {
	b, _ := json.Marshal(s)
	return template.JS(string(b))
}
