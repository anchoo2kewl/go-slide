package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/anchoo2kewl/go-slide/theme"
)

func newH(t *testing.T) (*ThemesHandler, *theme.Registry) {
	t.Helper()
	r := theme.NewRegistry()
	theme.RegisterBuiltins(r)
	return &ThemesHandler{Registry: r}, r
}

func do(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestThemes_List(t *testing.T) {
	h, _ := newH(t)
	w := do(t, h, "GET", "/", "")
	if w.Code != 200 {
		t.Fatalf("got %d", w.Code)
	}
	var out []map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &out)
	if len(out) < 2 {
		t.Errorf("expected ≥2 themes, got %d", len(out))
	}
}

func TestThemes_Get(t *testing.T) {
	h, _ := newH(t)
	w := do(t, h, "GET", "/aal", "")
	if w.Code != 200 {
		t.Fatalf("got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"id":"aal"`) {
		t.Errorf("expected aal theme in body, got %s", w.Body.String())
	}
}

func TestThemes_GetUnknown(t *testing.T) {
	h, _ := newH(t)
	w := do(t, h, "GET", "/nope", "")
	if w.Code != 404 {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestThemes_CreateAndDelete(t *testing.T) {
	h, _ := newH(t)
	body := `{"id":"midnight","name":"Midnight","css":".x{}","tokens":{"slide-bg":"#000"}}`
	w := do(t, h, "POST", "/", body)
	if w.Code != 201 {
		t.Fatalf("create: got %d, body=%s", w.Code, w.Body.String())
	}
	// Delete it
	w = do(t, h, "DELETE", "/midnight", "")
	if w.Code != 204 {
		t.Errorf("delete: got %d", w.Code)
	}
	// Should be gone
	w = do(t, h, "GET", "/midnight", "")
	if w.Code != 404 {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}

func TestThemes_BuiltinReadOnly(t *testing.T) {
	h, _ := newH(t)
	w := do(t, h, "PUT", "/aal", `{"name":"hax"}`)
	if w.Code != 403 {
		t.Errorf("expected 403 for PUT on built-in, got %d", w.Code)
	}
	w = do(t, h, "DELETE", "/aal", "")
	if w.Code != 403 {
		t.Errorf("expected 403 for DELETE on built-in, got %d", w.Code)
	}
}

func TestThemes_OnChangeFires(t *testing.T) {
	h, _ := newH(t)
	calls := []string{}
	h.OnChange = func(action string, _ theme.Theme) { calls = append(calls, action) }
	_ = do(t, h, "POST", "/", `{"id":"tk","name":"TK","css":".y{}"}`)
	_ = do(t, h, "PUT", "/tk", `{"name":"TK2","css":".y{}"}`)
	_ = do(t, h, "DELETE", "/tk", "")
	want := []string{"create", "update", "delete"}
	if strings.Join(calls, ",") != strings.Join(want, ",") {
		t.Errorf("OnChange calls = %v, want %v", calls, want)
	}
}
