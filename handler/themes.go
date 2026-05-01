package handler

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/anchoo2kewl/go-slide/theme"
)

// ThemesHandler exposes the theme registry over HTTP so an admin UI can
// list, create, update, and delete user-defined themes at runtime.
//
// Built-in themes (Default, AAL) are read-only — attempts to PUT/DELETE
// them return 403. Host applications using a database-backed registry
// should re-register loaded themes on startup.
//
// Routes (mount under any prefix):
//
//	GET    /            — list all themes
//	GET    /{id}        — get a single theme by ID
//	POST   /            — create a new theme (JSON body)
//	PUT    /{id}        — update an existing theme
//	DELETE /{id}        — delete a user theme
type ThemesHandler struct {
	Registry *theme.Registry
	// OnChange is called after a successful create/update/delete so the
	// host can persist the change to a database.
	OnChange func(action string, t theme.Theme)
}

// ServeHTTP implements http.Handler. It dispatches based on method and
// the trailing path segment.
func (h *ThemesHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := strings.Trim(r.URL.Path, "/")
	if i := strings.LastIndex(id, "/"); i >= 0 {
		id = id[i+1:]
	}

	switch r.Method {
	case http.MethodGet:
		if id == "" {
			h.list(w, r)
		} else {
			h.get(w, r, id)
		}
	case http.MethodPost:
		h.create(w, r)
	case http.MethodPut:
		h.update(w, r, id)
	case http.MethodDelete:
		h.delete(w, r, id)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// dto is the wire format. It mirrors theme.Theme but uses string for CSS
// so JSON input/output stays simple.
type dto struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	CSS             string            `json:"css"`
	Tokens          map[string]string `json:"tokens"`
	Fonts           []string          `json:"fonts"`
	BackgroundColor string            `json:"background_color"`
	TextColor       string            `json:"text_color"`
	Builtin         bool              `json:"builtin"`
}

func toDTO(t theme.Theme) dto {
	return dto{
		ID: t.ID, Name: t.Name, Description: t.Description,
		CSS: string(t.CSS), Tokens: t.Tokens, Fonts: t.Fonts,
		BackgroundColor: t.BackgroundColor, TextColor: t.TextColor,
		Builtin: t.Builtin,
	}
}

func fromDTO(d dto) theme.Theme {
	return theme.Theme{
		ID: d.ID, Name: d.Name, Description: d.Description,
		CSS: template.CSS(d.CSS), Tokens: d.Tokens, Fonts: d.Fonts,
		BackgroundColor: d.BackgroundColor, TextColor: d.TextColor,
		Builtin: false, // user-created themes are never built-in
	}
}

func (h *ThemesHandler) list(w http.ResponseWriter, r *http.Request) {
	all := h.Registry.All()
	out := make([]dto, len(all))
	for i, t := range all {
		out[i] = toDTO(t)
	}
	writeJSON(w, http.StatusOK, out)
}

func (h *ThemesHandler) get(w http.ResponseWriter, r *http.Request, id string) {
	t, ok := h.Registry.Get(id)
	if !ok {
		http.Error(w, "Theme not found", http.StatusNotFound)
		return
	}
	writeJSON(w, http.StatusOK, toDTO(t))
}

func (h *ThemesHandler) create(w http.ResponseWriter, r *http.Request) {
	var d dto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	if d.ID == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	if _, exists := h.Registry.Get(d.ID); exists {
		// .Get returns ok=true only on exact match
		if existing, ok := h.Registry.Get(d.ID); ok && existing.ID == d.ID {
			http.Error(w, fmt.Sprintf("theme %q already exists", d.ID), http.StatusConflict)
			return
		}
	}
	t := fromDTO(d)
	if err := h.Registry.Register(t); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if h.OnChange != nil {
		h.OnChange("create", t)
	}
	writeJSON(w, http.StatusCreated, toDTO(t))
}

func (h *ThemesHandler) update(w http.ResponseWriter, r *http.Request, id string) {
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}
	existing, ok := h.Registry.Get(id)
	if !ok || existing.ID != id {
		http.Error(w, "Theme not found", http.StatusNotFound)
		return
	}
	if existing.Builtin {
		http.Error(w, "Built-in themes are read-only", http.StatusForbidden)
		return
	}
	var d dto
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	d.ID = id // path param wins
	t := fromDTO(d)
	if err := h.Registry.Register(t); err != nil { // Register also overwrites
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if h.OnChange != nil {
		h.OnChange("update", t)
	}
	writeJSON(w, http.StatusOK, toDTO(t))
}

func (h *ThemesHandler) delete(w http.ResponseWriter, r *http.Request, id string) {
	existing, ok := h.Registry.Get(id)
	if !ok || existing.ID != id {
		http.Error(w, "Theme not found", http.StatusNotFound)
		return
	}
	if existing.Builtin {
		http.Error(w, "Built-in themes cannot be deleted", http.StatusForbidden)
		return
	}
	h.Registry.Unregister(id)
	if h.OnChange != nil {
		h.OnChange("delete", existing)
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
