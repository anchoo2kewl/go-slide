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
	"html/template"
	"net/http"

	"github.com/anchoo2kewl/go-slide/editor"
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

// ThemesHandler exposes the theme registry over HTTP for an admin UI.
//
// The host application is responsible for putting an auth middleware in
// front of this handler. Pass an onChange callback to persist user-created
// themes to a database; otherwise themes live only for the process's
// lifetime.
//
// Routes (relative to wherever you mount it):
//
//	GET    /            list themes
//	GET    /{id}        get one theme
//	POST   /            create a new theme (JSON body: id, name, css, ...)
//	PUT    /{id}        update an existing user theme
//	DELETE /{id}        delete a user theme
//
// Built-in themes (Default, AAL) are read-only.
func (e *Engine) ThemesHandler(onChange func(action string, t theme.Theme)) http.Handler {
	return &handler.ThemesHandler{Registry: e.themes, OnChange: onChange}
}

// EditorRequest configures one editor instance for an admin page.
type EditorRequest struct {
	// InitialContent is the existing slide HTML loaded on mount.
	InitialContent string
	// ThemeID picks which theme drives the canvas. Falls back to the
	// registry default when unknown.
	ThemeID string
	// AutosaveEndpoint is the POST URL the editor calls to autosave
	// (optional). The host is responsible for auth/persistence.
	AutosaveEndpoint string
	// OutputFieldID, OutputFieldName, FormID let the host wire the
	// editor into its existing <form>. Defaults: "go-slide-content",
	// "content", "slide-form".
	OutputFieldID   string
	OutputFieldName string
	FormID          string
}

// RenderEditor returns the HTML fragment that mounts the go-slide editor
// in an admin page. The host must serve the editor's assets under
// e.BasePath() + "/assets/" (use AssetHandler).
func (e *Engine) RenderEditor(req EditorRequest) (template.HTML, error) {
	th, _ := e.themes.Get(req.ThemeID)
	return editor.Render(editor.Config{
		AssetBase:        e.basePath + "/assets",
		PreviewEndpoint:  e.basePath + "/preview",
		AutosaveEndpoint: req.AutosaveEndpoint,
		Theme:            th,
		InitialContent:   req.InitialContent,
		OutputFieldID:    req.OutputFieldID,
		OutputFieldName:  req.OutputFieldName,
		FormID:           req.FormID,
	})
}
