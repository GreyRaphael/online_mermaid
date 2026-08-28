package server

import (
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

func spaHandler(assets fs.FS) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if name == "" || name == "." {
			name = "index.html"
		}

		f, err := assets.Open(name)
		if err == nil {
			stat, statErr := f.Stat()
			if statErr != nil {
				_ = f.Close()
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			if !stat.IsDir() {
				serveFile(w, r, f, name)
				return
			}
			_ = f.Close()
		} else if !errors.Is(err, fs.ErrNotExist) {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Fallback to index.html for SPA client-side routes.
		indexFile, err := assets.Open("index.html")
		if err != nil {
			http.Error(w, "SPA index.html not found", http.StatusNotFound)
			return
		}
		serveFile(w, r, indexFile, "index.html")
	})
}

func serveFile(w http.ResponseWriter, r *http.Request, f fs.File, name string) {
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if stat.IsDir() {
		http.NotFound(w, r)
		return
	}

	if ext := filepath.Ext(name); ext != "" {
		if mimeType := mime.TypeByExtension(ext); mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		}
	}
	if strings.HasPrefix(name, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}

	readSeeker, ok := f.(io.ReadSeeker)
	if !ok {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, name, stat.ModTime(), readSeeker)
}
