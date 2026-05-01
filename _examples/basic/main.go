// A minimal go-slide example. Run:
//
//	go run ./_examples/basic
//
// Then open http://localhost:8090/slides/demo
package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	goslide "github.com/anchoo2kewl/go-slide"
	"github.com/anchoo2kewl/go-slide/theme"
)

func main() {
	registry := theme.NewRegistry()
	theme.RegisterBuiltins(registry)

	engine, err := goslide.New(
		goslide.WithThemes(registry),
		goslide.WithBasePath("/slides"),
	)
	if err != nil {
		log.Fatal(err)
	}

	demo := goslide.Slide{
		Title:          "Demo Deck",
		Slug:           "demo",
		ContentHTML:    template.HTML(demoHTML),
		ThemeID:        "aal",
		AuthorName:     "Anshuman Biswas",
		AuthorUsername: "anchoo2kewl",
		Coauthors: []goslide.Coauthor{
			{Name: "Gary Zeng", Username: "gary"},
		},
		BackURL: "/",
	}

	http.HandleFunc("/slides/demo", func(w http.ResponseWriter, r *http.Request) {
		out, err := engine.RenderPresentation(demo)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, out)
	})

	http.Handle("/slides/preview", engine.PreviewHandler())
	http.Handle("/slides/assets/", http.StripPrefix("/slides/assets/", engine.AssetHandler()))

	log.Println("listening on http://localhost:8090/slides/demo")
	log.Fatal(http.ListenAndServe(":8090", nil))
}

const demoHTML = `
<section class="aal pptx-slide" data-background-color="#0A1628">
  <div class="aal-wrap">
    <div class="aal-bar"></div>
    <span class="aal-eyebrow">Demo · 2026</span>
    <h1 class="aal-h1" style="margin-top:1.2rem">go-slide</h1>
    <hr class="aal-rule"/>
    <p class="aal-lede">An embeddable slide-deck engine for Go web apps.</p>
    <div style="margin-top:2rem;display:flex;gap:.6rem;flex-wrap:wrap">
      <span class="aal-pill">Themes</span>
      <span class="aal-pill">TipTap WYSIWYG</span>
      <span class="aal-pill">Reveal.js</span>
    </div>
  </div>
</section>

<section class="aal-light pptx-slide" data-background-color="#F0F4F8">
  <div class="aal-wrap">
    <span class="aal-eyebrow-light">Why</span>
    <h2 class="aal-h2" style="margin-top:1rem">One source of truth.</h2>
    <hr class="aal-rule"/>
    <div class="aal-grid-3" style="margin-top:1.6rem">
      <div class="aal-card-light">
        <div class="aal-stat-light">1</div>
        <div class="aal-stat-label">Editor canvas, preview iframe, and prod page all load the same theme CSS.</div>
      </div>
      <div class="aal-card-light">
        <div class="aal-stat-light">2</div>
        <div class="aal-stat-label">TipTap schema knows every design primitive — WYSIWYG round-trips losslessly.</div>
      </div>
      <div class="aal-card-light">
        <div class="aal-stat-light">3</div>
        <div class="aal-stat-label">Themes register at runtime so users can create and manage their own.</div>
      </div>
    </div>
  </div>
</section>
`
