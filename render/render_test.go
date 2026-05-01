package render

import (
	"html/template"
	"strings"
	"testing"

	"github.com/anchoo2kewl/go-slide/theme"
)

func newTestRenderer(t *testing.T) *Renderer {
	t.Helper()
	r := theme.NewRegistry()
	theme.RegisterBuiltins(r)
	rr, err := New(Options{Themes: r})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return rr
}

func TestRenderPresentation_AALTheme(t *testing.T) {
	r := newTestRenderer(t)
	html, err := r.RenderPresentationString(Slide{
		Title:       "Demo",
		ContentHTML: template.HTML(`<section class="aal pptx-slide"><div class="aal-wrap"><h1 class="aal-h1">Hi</h1></div></section>`),
		ThemeID:     "aal",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{
		"reveal.js",
		"<title>Demo</title>",
		`class="aal pptx-slide"`,
		"aal-h1",
		"--slide-accent: #1DB89E",
		"Reveal.initialize",
		"print-pdf",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered HTML missing %q", want)
		}
	}
}

func TestRenderPresentation_DefaultTheme(t *testing.T) {
	r := newTestRenderer(t)
	html, err := r.RenderPresentationString(Slide{
		Title:       "Talk",
		ContentHTML: template.HTML(`<section class="slide pptx-slide"><div class="slide-wrap"><h1 class="slide-h1">Talk</h1></div></section>`),
		ThemeID:     "default",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{
		"slide-h1",
		"--slide-accent: #2563eb",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered HTML missing %q", want)
		}
	}
}

func TestRenderPresentation_UnknownThemeFallsBack(t *testing.T) {
	r := newTestRenderer(t)
	html, err := r.RenderPresentationString(Slide{
		Title:       "Talk",
		ContentHTML: template.HTML(`<section><h1>Hi</h1></section>`),
		ThemeID:     "no-such-theme",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	// Default is the fallback (RegisterBuiltins sets it).
	if !strings.Contains(html, "--slide-accent: #2563eb") {
		t.Errorf("expected default theme to be applied as fallback")
	}
}

func TestRenderPresentation_AuthorAndCoauthors(t *testing.T) {
	r := newTestRenderer(t)
	html, err := r.RenderPresentationString(Slide{
		Title:           "Pitch",
		ContentHTML:     template.HTML(`<section class="aal pptx-slide"><h1>Hi</h1></section>`),
		ThemeID:         "aal",
		AuthorName:      "Anshuman",
		AuthorUsername:  "anchoo2kewl",
		AuthorAvatarURL: "/u/a.jpg",
		Coauthors: []Coauthor{
			{Name: "Gary", Username: "gary", AvatarURL: "/u/g.jpg"},
		},
		EditURL: "/admin/slides/1/edit",
	})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{
		`title="Anshuman (author)"`,
		`title="Gary (co-author)"`,
		`href="/admin/slides/1/edit"`,
		`href="?print-pdf"`,
	} {
		if !strings.Contains(html, want) {
			t.Errorf("rendered HTML missing %q", want)
		}
	}
}

func TestNew_RequiresThemes(t *testing.T) {
	if _, err := New(Options{}); err == nil {
		t.Errorf("expected error when Themes is nil")
	}
}
