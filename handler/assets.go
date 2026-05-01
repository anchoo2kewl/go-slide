package handler

import (
	"net/http"

	"github.com/anchoo2kewl/go-slide/editor"
)

// AssetHandler serves the editor's embedded CSS / JS bundle.
// Mount it at "<basePath>/assets/" with http.StripPrefix.
func AssetHandler() http.Handler {
	return http.FileServer(http.FS(editor.Assets()))
}
