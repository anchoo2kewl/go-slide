package goslide

import "github.com/anchoo2kewl/go-slide/theme"

// Option configures an Engine constructed via New.
type Option func(*Engine)

// WithThemes is required: the registry that holds available themes.
func WithThemes(r *theme.Registry) Option {
	return func(e *Engine) { e.themes = r }
}

// WithBasePath sets the URL prefix the engine is mounted under.
// Default: "/slides".
func WithBasePath(p string) Option {
	return func(e *Engine) {
		if p != "" {
			e.basePath = p
		}
	}
}

// WithRevealCDN overrides the reveal.js CDN base URL. Useful for air-gapped
// deployments that ship reveal locally.
// Default: "https://cdn.jsdelivr.net/npm/reveal.js@4.3.1".
func WithRevealCDN(url string) Option {
	return func(e *Engine) { e.cdn = url }
}
