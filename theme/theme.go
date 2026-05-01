// Package theme defines the visual design system for slides.
//
// A Theme is a self-contained bundle of CSS, design tokens, and font imports
// that the slide canvas, the editor preview, and the rendered presentation
// page all share. Because all three contexts load the same theme CSS, the
// "what you see is what you get" promise is enforced at the source: there is
// only one stylesheet defining each visual primitive.
//
// Themes are registered at runtime via Registry. The host application can
// register the built-in themes (default, AAL) and/or user-defined themes
// loaded from a database, files, or a CMS.
package theme

import (
	"fmt"
	"html/template"
	"sort"
	"sync"
)

// Theme is one visual design system for slides.
//
// CSS is the bundle that styles slide content. It MUST be self-contained
// (no external @imports beyond Fonts). It is loaded by:
//   1. the rendered presentation page (production)
//   2. the editor's WYSIWYG canvas
//   3. the editor's preview iframe
//
// Tokens are CSS custom properties published on :root. They let host pages
// or other themes reference shared values (--accent, --bg, --text, etc.).
type Theme struct {
	// ID is the stable slug used to reference the theme. Lowercase, kebab-case.
	ID string
	// Name is the human-readable display name.
	Name string
	// Description is a short blurb shown in theme pickers.
	Description string
	// CSS is the bundled stylesheet for slide content.
	CSS template.CSS
	// Tokens are CSS custom properties exposed on :root. Optional.
	Tokens map[string]string
	// Fonts are full <link> URLs (Google Fonts, etc.) loaded before the CSS.
	Fonts []string
	// BackgroundColor is the default slide background when a section
	// doesn't set its own data-background-color.
	BackgroundColor string
	// TextColor is the default body text color.
	TextColor string
	// Builtin marks themes shipped by go-slide so the admin UI can
	// distinguish them from user-created ones.
	Builtin bool
}

// RootCSS returns a <style>-ready string with the theme tokens published
// as CSS custom properties on :root, followed by the theme's CSS bundle.
func (t Theme) RootCSS() template.CSS {
	if len(t.Tokens) == 0 {
		return t.CSS
	}
	keys := make([]string, 0, len(t.Tokens))
	for k := range t.Tokens {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := ":root {\n"
	for _, k := range keys {
		out += fmt.Sprintf("  --%s: %s;\n", k, t.Tokens[k])
	}
	out += "}\n\n"
	out += string(t.CSS)
	return template.CSS(out)
}

// FontTags renders the theme's font <link> tags as HTML.
func (t Theme) FontTags() template.HTML {
	if len(t.Fonts) == 0 {
		return ""
	}
	out := ""
	for _, url := range t.Fonts {
		// Best-effort sanitization: themes are trusted code, but
		// quotes must not break the attribute.
		safe := template.HTMLEscapeString(url)
		out += fmt.Sprintf(`<link rel="stylesheet" href="%s">`+"\n", safe)
	}
	return template.HTML(out)
}

// Registry stores themes addressable by ID. It is safe for concurrent use.
type Registry struct {
	mu       sync.RWMutex
	themes   map[string]Theme
	fallback string
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{themes: make(map[string]Theme)}
}

// Register adds or replaces a theme. Returns an error if the theme has no ID.
func (r *Registry) Register(t Theme) error {
	if t.ID == "" {
		return fmt.Errorf("theme: ID is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.themes[t.ID] = t
	if r.fallback == "" {
		r.fallback = t.ID
	}
	return nil
}

// MustRegister is Register but panics on error. Convenient at init time.
func (r *Registry) MustRegister(t Theme) {
	if err := r.Register(t); err != nil {
		panic(err)
	}
}

// Unregister removes a theme by ID. No-op if not present.
func (r *Registry) Unregister(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.themes, id)
	if r.fallback == id {
		r.fallback = ""
		for k := range r.themes {
			r.fallback = k
			break
		}
	}
}

// Get returns the theme by ID. If not found, returns the fallback theme
// and ok=false. If the registry is empty, returns the zero Theme and ok=false.
func (r *Registry) Get(id string) (Theme, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, ok := r.themes[id]; ok {
		return t, true
	}
	if r.fallback != "" {
		return r.themes[r.fallback], false
	}
	return Theme{}, false
}

// SetFallback chooses which theme to return when Get is called for an
// unknown ID. The chosen theme must already be registered.
func (r *Registry) SetFallback(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.themes[id]; !ok {
		return fmt.Errorf("theme: cannot set fallback to unregistered theme %q", id)
	}
	r.fallback = id
	return nil
}

// All returns every registered theme, sorted by ID.
func (r *Registry) All() []Theme {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Theme, 0, len(r.themes))
	for _, t := range r.themes {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
