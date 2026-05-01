# go-slide

Embeddable slide-deck engine for Go web apps. Built on [reveal.js](https://revealjs.com/) for presentation chrome and [TipTap](https://tiptap.dev/) for WYSIWYG editing, with a runtime-registerable theme system so a single Go binary can serve multiple visual identities.

```go
import (
    goslide "github.com/anchoo2kewl/go-slide"
    "github.com/anchoo2kewl/go-slide/theme"
)

registry := theme.NewRegistry()
theme.RegisterBuiltins(registry)              // Default + AAL

engine, _ := goslide.New(
    goslide.WithThemes(registry),
    goslide.WithBasePath("/slides"),
)

http.Handle("/slides/preview", engine.PreviewHandler())
http.Handle("/slides/assets/", http.StripPrefix("/slides/assets/", engine.AssetHandler()))
http.HandleFunc("/slides/demo", func(w http.ResponseWriter, r *http.Request) {
    html, _ := engine.RenderPresentation(goslide.Slide{
        Title: "Demo", ContentHTML: template.HTML("<section>...</section>"),
        ThemeID: "aal",
    })
    fmt.Fprint(w, html)
})
```

See `_examples/basic` for a runnable demo.

## Features

- **Reveal.js presentation page** — fade transitions, controls, slide numbers, fullscreen toggle, keyboard nav.
- **One-click PDF download** — `?print-pdf` query param triggers reveal's print mode and auto-fires `window.print()`. No server-side headless browser required.
- **Theme system** — `theme.Registry` with built-in `Default` (clean white) and `AAL` (dark navy + teal). User-defined themes register at runtime.
- **WYSIWYG editor** *(in progress)* — TipTap with custom Node extensions for every design primitive. The editor canvas, the preview iframe, and the production page all load the same theme CSS, so what you see is what you get. *Day-2 work.*
- **Code mode** *(in progress)* — CodeMirror with HTML highlighting for raw editing. *Day-2 work.*

## Why "single source of truth"?

Slide editors traditionally have three different stylesheets:
1. The editor's WYSIWYG canvas styles.
2. The editor's preview iframe styles.
3. The actual rendered presentation styles.

When those drift, authors see one thing in the editor and a different thing on the deployed page. go-slide flips this: the theme CSS is the source of truth, loaded by all three contexts. There is no separate "editor look" — typing in the editor produces the exact pixels that ship.

## Themes

A theme is a self-contained CSS bundle plus design tokens:

```go
type Theme struct {
    ID, Name, Description string
    CSS template.CSS         // the stylesheet
    Tokens map[string]string // CSS custom properties on :root
    Fonts []string           // Google Fonts URLs
    BackgroundColor string
    TextColor string
    Builtin bool
}
```

Built-ins:

| ID | Use |
|----|----|
| `default` | Clean white deck for general talks. |
| `aal`     | Dark navy + teal pitch theme (AI Agent Lens). |

User themes:

```go
registry.Register(theme.Theme{
    ID: "midnight",
    Name: "Midnight Blue",
    CSS: template.CSS(myCSS),
    Tokens: map[string]string{
        "slide-bg": "#0c1844",
        "slide-accent": "#fcb045",
    },
    Fonts: []string{"https://fonts.googleapis.com/css2?family=Manrope:wght@300;700&display=swap"},
})
```

The host app can load themes from a database, a file watcher, or a CMS — any time before or after `goslide.New()`.

## Module layout

```
slide.go        # Engine type, New(opts...), public API
options.go      # WithThemes, WithBasePath, WithRevealCDN, ...
theme/          # Theme + Registry, built-in themes, embedded CSS
render/         # Renderer, present.gohtml template
handler/        # PreviewHandler, AssetHandler
editor/         # CSS, JS, templates for the embeddable editor (Day-2)
_examples/      # runnable demos
```

Mirrors `go-wiki` and `go-draw` so the integration pattern is familiar.

## Status

Day-1 scaffold (renderer + theme system + present page) is in place. Day-2 work adds the TipTap design-system schema, the WYSIWYG canvas, and the preview wiring.

## License

MIT — see [LICENSE](LICENSE).
