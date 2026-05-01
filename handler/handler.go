// Package handler exposes HTTP handlers that can be mounted into any
// chi/gorilla/net-http app.
package handler

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/anchoo2kewl/go-slide/render"
)

// PreviewHandler renders a slide's HTML for live preview in the editor.
//
// Accepts POST with form fields:
//
//	content    — raw <section>... HTML
//	theme      — theme ID (optional)
//	title      — slide title (optional)
//
// Returns: {"html": "<full reveal page>"}
type PreviewHandler struct {
	Renderer *render.Renderer
}

// NewPreviewHandler creates a PreviewHandler for the given renderer.
func NewPreviewHandler(r *render.Renderer) *PreviewHandler {
	return &PreviewHandler{Renderer: r}
}

// ServeHTTP implements http.Handler.
func (h *PreviewHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form", http.StatusBadRequest)
		return
	}
	s := render.Slide{
		Title:       r.FormValue("title"),
		ContentHTML: template.HTML(r.FormValue("content")),
		ThemeID:     r.FormValue("theme"),
	}
	out, err := h.Renderer.RenderPresentationString(s)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"html": out})
}
