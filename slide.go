// Package goslide is an embeddable slide-deck engine for Go web apps.
//
// It provides:
//
//   - a renderer that produces full reveal.js presentation pages from slide
//     HTML and a chosen theme
//   - a built-in theme system (Default, AAL) plus a runtime registry for
//     user-created themes
//   - HTTP handlers for live preview and asset serving
//   - an embedded TipTap-based editor whose canvas, preview, and the
//     production presentation page all share one stylesheet so "what you
//     see is what you get" is enforced at the source
//
// Usage:
//
//	r := theme.NewRegistry()
//	theme.RegisterBuiltins(r)
//
//	s, err := goslide.New(
//	  goslide.WithThemes(r),
//	  goslide.WithBasePath("/slides"),
//	)
//
//	html, _ := s.RenderPresentation(slide)            // string of full HTML
//	http.Handle("/slides/preview",  s.PreviewHandler())
//	http.Handle("/slides/assets/", http.StripPrefix("/slides/assets/", s.AssetHandler()))
package goslide

import (
	"fmt"
	"net/http"

	"github.com/anchoo2kewl/go-slide/handler"
	"github.com/anchoo2kewl/go-slide/render"
	"github.com/anchoo2kewl/go-slide/theme"
)

// Slide is re-exported here so callers don't need to import the render package
// for the common case of rendering a single deck.
type Slide = render.Slide

// Coauthor mirrors render.Coauthor for the same reason.
type Coauthor = render.Coauthor

// Engine is the top-level handle that ties together the theme registry, the
// renderer, and the HTTP handlers.
type Engine struct {
	themes   *theme.Registry
	renderer *render.Renderer
	basePath string
	cdn      string
}

// New constructs an Engine. WithThemes is required.
func New(opts ...Option) (*Engine, error) {
	e := &Engine{basePath: "/slides"}
	for _, o := range opts {
		o(e)
	}
	if e.themes == nil {
		return nil, fmt.Errorf("goslide: WithThemes(*theme.Registry) is required")
	}
	r, err := render.New(render.Options{
		Themes:    e.themes,
		RevealCDN: e.cdn,
	})
	if err != nil {
		return nil, err
	}
	e.renderer = r
	return e, nil
}

// Themes exposes the registry so callers can register additional themes
// (e.g. loaded from a database) after construction.
func (e *Engine) Themes() *theme.Registry { return e.themes }

// Renderer exposes the renderer for advanced uses.
func (e *Engine) Renderer() *render.Renderer { return e.renderer }

// BasePath returns the configured base path (e.g. "/slides").
func (e *Engine) BasePath() string { return e.basePath }

// RenderPresentation returns the full HTML for a presentation page.
func (e *Engine) RenderPresentation(s Slide) (string, error) {
	return e.renderer.RenderPresentationString(s)
}

// PreviewHandler accepts a POST with form fields:
//
//	content   — raw <section>... HTML
//	theme     — theme ID (optional; falls back to registry default)
//
// and returns {"html":"..."} JSON containing the rendered presentation.
func (e *Engine) PreviewHandler() http.Handler {
	return handler.NewPreviewHandler(e.renderer)
}

// AssetHandler serves the embedded CSS / JS bundle for the editor.
// Mount it at e.BasePath() + "/assets/".
func (e *Engine) AssetHandler() http.Handler {
	return handler.AssetHandler()
}
