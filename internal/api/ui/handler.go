package ui

import (
	"fmt"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

const apiPath = "/api"

type handler struct {
	fs        fs.FS
	files     http.Handler
	indexHTML []byte
}

// New returns a static file handler with an index fallback for SPA routes.
func New(files fs.FS) (http.Handler, error) {
	indexHTML, err := fs.ReadFile(files, "index.html")
	if err != nil {
		return nil, fmt.Errorf("read UI index: %w", err)
	}

	return &handler{fs: files, files: http.FileServerFS(files), indexHTML: indexHTML}, nil
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	if r.URL.Path == apiPath || strings.HasPrefix(r.URL.Path, apiPath+"/") {
		http.NotFound(w, r)
		return
	}

	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name != "." {
		if info, err := fs.Stat(h.fs, name); err == nil && !info.IsDir() {
			if strings.HasPrefix(r.URL.Path, "/_app/immutable/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			h.files.ServeHTTP(w, r)
			return
		}
		if path.Ext(name) != "" {
			http.NotFound(w, r)
			return
		}
	}

	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Content-Type", mime.TypeByExtension(".html"))
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodGet {
		_, _ = w.Write(h.indexHTML)
	}
}
