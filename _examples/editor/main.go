// go-slide editor demo. Boots a tiny HTTP server that mounts the editor
// at /edit and the slide page at /slide so you can compare WYSIWYG /
// Code / Preview against the production renderer side-by-side.
//
//	go run ./_examples/editor
//
// Then open http://localhost:8091/edit
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
	r := theme.NewRegistry()
	theme.RegisterBuiltins(r)

	engine, err := goslide.New(
		goslide.WithThemes(r),
		goslide.WithBasePath("/slides"),
	)
	if err != nil {
		log.Fatal(err)
	}

	demo := goslide.Slide{
		Title:          "Editor Demo",
		Slug:           "demo",
		ContentHTML:    template.HTML(demoHTML),
		ThemeID:        "aal",
		AuthorName:     "Anshuman Biswas",
		AuthorUsername: "anchoo2kewl",
	}

	// Engine routes
	http.Handle("/slides/preview", engine.PreviewHandler())
	http.Handle("/slides/assets/", http.StripPrefix("/slides/assets", engine.AssetHandler()))

	// /slide — production-style render of the demo deck
	http.HandleFunc("/slide", func(w http.ResponseWriter, _ *http.Request) {
		out, err := engine.RenderPresentation(demo)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, out)
	})

	// /edit — page that mounts the editor at full size.
	// Note we use io.WriteString rather than Fprintf so percent signs
	// inside the editor's embedded CSS aren't interpreted as format verbs.
	http.HandleFunc("/edit", func(w http.ResponseWriter, _ *http.Request) {
		editor, err := engine.RenderEditor(goslide.EditorRequest{
			InitialContent: demoHTML,
			ThemeID:        "aal",
		})
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		const head = `<!DOCTYPE html><html><head>
<meta charset="utf-8"/>
<title>Editor Demo</title>
<style>html,body{margin:0;padding:0;height:100%;background:#0f172a;color:#e2e8f0;font-family:Inter,sans-serif}
.topbar{height:64px;display:flex;align-items:center;padding:0 16px;background:#1e293b;border-bottom:1px solid #334155}</style>
</head><body>
<div class="topbar">go-slide editor demo</div>
`
		const tail = `</body></html>`
		fmt.Fprint(w, head)
		fmt.Fprint(w, string(editor))
		fmt.Fprint(w, tail)
	})

	log.Println("listening on http://localhost:8091/edit")
	log.Fatal(http.ListenAndServe(":8091", nil))
}

const demoHTML = `<section class="aal pptx-slide" data-background-color="#0A1628">
  <div class="aal-wrap">
    <div class="aal-bar"></div>
    <span class="aal-eyebrow">Seed Pitch · 2026</span>
    <h1 class="aal-h1" style="margin-top:1.2rem">AI Agent Lens</h1>
    <hr class="aal-rule"/>
    <p class="aal-lede">Runtime security &amp; compliance for the AI agents already running inside your company.</p>
    <div style="margin-top:2rem;display:flex;gap:.6rem;flex-wrap:wrap">
      <span class="aal-pill">AgentShield · OSS</span>
      <span class="aal-pill">AI Agent Lens · SaaS</span>
      <span class="aal-pill">Data Protection · DLP</span>
    </div>
  </div>
  <div class="aal-foot"><span>aiagentlens.com</span><span>CONFIDENTIAL</span></div>
</section>

<section class="aal-light pptx-slide" data-background-color="#F0F4F8">
  <div class="aal-wrap">
    <span class="aal-eyebrow-light">By the Numbers</span>
    <h2 class="aal-h2" style="margin-top:1rem">The blast radius is already here.</h2>
    <hr class="aal-rule"/>
    <div class="aal-grid-3" style="margin-top:1.6rem">
      <div class="aal-card-light">
        <div class="aal-stat-light">31</div>
        <div class="aal-stat-label">Credential files leaked</div>
      </div>
      <div class="aal-card-light">
        <div class="aal-stat-light">20+</div>
        <div class="aal-stat-label">Cloud / DB credential paths</div>
      </div>
      <div class="aal-card-light">
        <div class="aal-stat-light">10–30ms</div>
        <div class="aal-stat-label">End-to-end pipeline latency</div>
      </div>
    </div>
  </div>
</section>`
