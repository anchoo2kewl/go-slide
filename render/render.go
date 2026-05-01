// Package render produces the HTML for a slide presentation page.
//
// The renderer takes a Slide (raw section HTML + chosen theme) and produces
// a complete reveal.js page that:
//
//   - loads reveal.js core and the white chrome theme
//   - loads the selected go-slide theme's fonts and CSS
//   - embeds the slide sections verbatim
//   - initializes Reveal at 1280x800 with print-pdf auto-detection
//   - renders a top-right nav (Edit, Back, PDF, Fullscreen) when configured
//
// The same template is used by the production presentation page, by the
// editor's preview iframe, and by external embeds, so the visual output
// is identical in every context.
package render

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/anchoo2kewl/go-slide/theme"
)

//go:embed templates/*.gohtml
var templates embed.FS

// Slide is the data the renderer needs to produce a presentation page.
type Slide struct {
	Title       string
	Slug        string
	Description string
	// ContentHTML is the raw <section>...</section> sequence.
	ContentHTML template.HTML
	// ThemeID selects which theme to apply. If unknown, the registry
	// fallback is used (see theme.Registry.SetFallback).
	ThemeID string
	// AuthorName / AuthorAvatarURL / AuthorUsername populate the nav pill.
	AuthorName      string
	AuthorUsername  string
	AuthorAvatarURL string
	// Coauthors are additional contributors shown next to the author.
	Coauthors []Coauthor
	// EditURL — when non-empty, the nav shows an Edit button.
	EditURL string
	// BackURL — defaults to "/slides" when empty.
	BackURL string
	// CanonicalURL is used for og:url / canonical link.
	CanonicalURL string
}

// Coauthor is one secondary contributor.
type Coauthor struct {
	Name      string
	Username  string
	AvatarURL string
}

// Renderer renders Slide values to full HTML pages.
type Renderer struct {
	tpl    *template.Template
	themes *theme.Registry
	cdn    string // reveal.js base URL
}

// Options configure the renderer.
type Options struct {
	// Themes is the registry the renderer reads from. Required.
	Themes *theme.Registry
	// RevealCDN overrides the reveal.js asset base URL.
	// Default: "https://cdn.jsdelivr.net/npm/reveal.js@4.3.1".
	RevealCDN string
}

// New creates a Renderer.
func New(opts Options) (*Renderer, error) {
	if opts.Themes == nil {
		return nil, fmt.Errorf("render: Themes registry is required")
	}
	tpl, err := template.New("present").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
		"safeCSS":  func(s template.CSS) template.CSS { return s },
	}).ParseFS(templates, "templates/*.gohtml")
	if err != nil {
		return nil, fmt.Errorf("render: parse templates: %w", err)
	}
	cdn := opts.RevealCDN
	if cdn == "" {
		cdn = "https://cdn.jsdelivr.net/npm/reveal.js@4.3.1"
	}
	return &Renderer{tpl: tpl, themes: opts.Themes, cdn: cdn}, nil
}

// RenderPresentation writes the full presentation HTML to w.
func (r *Renderer) RenderPresentation(w io.Writer, s Slide) error {
	th, _ := r.themes.Get(s.ThemeID)
	data := struct {
		Slide     Slide
		Theme     theme.Theme
		ThemeCSS  template.CSS
		FontTags  template.HTML
		RevealCDN string
	}{
		Slide:     s,
		Theme:     th,
		ThemeCSS:  th.RootCSS(),
		FontTags:  th.FontTags(),
		RevealCDN: r.cdn,
	}
	return r.tpl.ExecuteTemplate(w, "present.gohtml", data)
}

// RenderPresentationString is a convenience wrapper returning the result as a string.
func (r *Renderer) RenderPresentationString(s Slide) (string, error) {
	var buf bytes.Buffer
	if err := r.RenderPresentation(&buf, s); err != nil {
		return "", err
	}
	return buf.String(), nil
}
