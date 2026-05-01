package theme

import (
	_ "embed"
	"html/template"
)

//go:embed builtin/default.css
var defaultCSS string

//go:embed builtin/aal.css
var aalCSS string

// Default is the neutral theme used for general-purpose decks.
//
// It targets a clean white background with crisp typography, suitable for
// engineering talks, retros, and content that doesn't want a strong brand
// identity. Reveal.js's white theme provides the chrome; Default adds the
// design-system primitives (cards, grids, pills) on top.
var Default = Theme{
	ID:              "default",
	Name:            "Default",
	Description:     "Clean white deck with neutral typography. Good for talks and retros.",
	CSS:             template.CSS(defaultCSS),
	BackgroundColor: "#ffffff",
	TextColor:       "#0f172a",
	Fonts: []string{
		"https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&display=swap",
	},
	Tokens: map[string]string{
		"slide-bg":        "#ffffff",
		"slide-text":      "#0f172a",
		"slide-text-soft": "#475569",
		"slide-accent":    "#2563eb",
		"slide-rule":      "#2563eb",
		"slide-card-bg":   "#f8fafc",
		"slide-card-bd":   "#e2e8f0",
	},
	Builtin: true,
}

// AAL is the AI Agent Lens pitch theme: dark navy with teal accents.
//
// Intended for investor decks and product overviews that need to feel
// premium and confident. Slides default to the dark backdrop; alternating
// "aal-light" sections use the soft gray for visual rhythm.
var AAL = Theme{
	ID:              "aal",
	Name:            "AI Agent Lens",
	Description:     "Dark navy + teal accent system used for investor and product decks.",
	CSS:             template.CSS(aalCSS),
	BackgroundColor: "#0A1628",
	TextColor:       "#F0F4F8",
	Fonts: []string{
		"https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;600&display=swap",
	},
	Tokens: map[string]string{
		"slide-bg":         "#0A1628",
		"slide-bg-light":   "#F0F4F8",
		"slide-text":       "#F0F4F8",
		"slide-text-soft":  "#CFD8DC",
		"slide-text-mute":  "#90A4AE",
		"slide-accent":     "#1DB89E",
		"slide-accent-2":   "#0E8C7D",
		"slide-rule":       "#1DB89E",
		"slide-card-bg":    "rgba(255,255,255,0.04)",
		"slide-card-bd":    "rgba(29,184,158,0.28)",
	},
	Builtin: true,
}

// RegisterBuiltins registers Default and AAL on the given Registry, with
// Default as the fallback. Call this once at startup.
func RegisterBuiltins(r *Registry) {
	r.MustRegister(Default)
	r.MustRegister(AAL)
	_ = r.SetFallback(Default.ID)
}
