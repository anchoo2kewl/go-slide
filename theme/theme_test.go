package theme

import (
	"strings"
	"testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(Theme{ID: "x", Name: "X"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	got, ok := r.Get("x")
	if !ok || got.Name != "X" {
		t.Errorf("Get('x') = (%v, %v), want X/true", got, ok)
	}
}

func TestRegistry_RegisterRequiresID(t *testing.T) {
	r := NewRegistry()
	if err := r.Register(Theme{Name: "no-id"}); err == nil {
		t.Errorf("expected error for empty ID")
	}
}

func TestRegistry_FallbackOnUnknown(t *testing.T) {
	r := NewRegistry()
	r.MustRegister(Theme{ID: "a", Name: "A"})
	r.MustRegister(Theme{ID: "b", Name: "B"})
	_ = r.SetFallback("a")
	got, ok := r.Get("nope")
	if ok {
		t.Errorf("expected ok=false for unknown ID")
	}
	if got.ID != "a" {
		t.Errorf("expected fallback A, got %q", got.ID)
	}
}

func TestRegistry_AllSorted(t *testing.T) {
	r := NewRegistry()
	r.MustRegister(Theme{ID: "z"})
	r.MustRegister(Theme{ID: "a"})
	r.MustRegister(Theme{ID: "m"})
	all := r.All()
	if len(all) != 3 || all[0].ID != "a" || all[2].ID != "z" {
		t.Errorf("All() = %v, want sorted", all)
	}
}

func TestTheme_RootCSSEmbedsTokens(t *testing.T) {
	tt := Theme{
		ID:  "tk",
		CSS: ".x { color: red; }",
		Tokens: map[string]string{
			"primary": "#000",
			"accent":  "#f00",
		},
	}
	out := string(tt.RootCSS())
	if !strings.Contains(out, "--accent: #f00") {
		t.Errorf("expected --accent token in RootCSS, got %q", out)
	}
	if !strings.Contains(out, ".x { color: red; }") {
		t.Errorf("expected base CSS preserved")
	}
	// Tokens should be sorted
	if strings.Index(out, "--accent") > strings.Index(out, "--primary") {
		t.Errorf("expected sorted tokens (--accent before --primary)")
	}
}

func TestTheme_RootCSSWithoutTokens(t *testing.T) {
	tt := Theme{ID: "tk", CSS: ".y {}"}
	if string(tt.RootCSS()) != ".y {}" {
		t.Errorf("expected RootCSS to return CSS unchanged when no tokens")
	}
}

func TestRegisterBuiltins(t *testing.T) {
	r := NewRegistry()
	RegisterBuiltins(r)
	if _, ok := r.Get("default"); !ok {
		t.Errorf("default theme not registered")
	}
	if _, ok := r.Get("aal"); !ok {
		t.Errorf("aal theme not registered")
	}
	// Default should be the fallback
	got, _ := r.Get("nonexistent")
	if got.ID != "default" {
		t.Errorf("expected default fallback, got %q", got.ID)
	}
}
