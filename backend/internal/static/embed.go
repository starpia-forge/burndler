package static

import (
	"embed"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"
)

//go:embed all:dist
var frontendFS embed.FS

// GetFrontendFS returns the embedded frontend filesystem
func GetFrontendFS() (fs.FS, error) {
	// Get the dist subdirectory from the embedded filesystem
	frontendSubFS, err := fs.Sub(frontendFS, "dist")
	if err != nil {
		return nil, err
	}
	return frontendSubFS, nil
}

// StaticFileHandler creates an HTTP handler for serving static files
func StaticFileHandler() (http.Handler, error) {
	frontendSubFS, err := GetFrontendFS()
	if err != nil {
		return nil, err
	}

	return http.FileServer(http.FS(frontendSubFS)), nil
}

// SPAHandler creates a handler that serves index.html for SPA routing
func SPAHandler() (http.HandlerFunc, error) {
	frontendSubFS, err := GetFrontendFS()
	if err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try to serve the file
		data, err := fs.ReadFile(frontendSubFS, path)
		if err != nil {
			// File not found, serve index.html for SPA routing
			indexData, indexErr := fs.ReadFile(frontendSubFS, "index.html")
			if indexErr != nil {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			// Set content type for HTML
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			if _, err := w.Write(indexData); err != nil {
				// Log error but don't return error response as headers are already sent
				// This follows standard HTTP handler patterns
				_ = err // Explicitly ignore error as response cannot be changed
			}
			return
		}

		// Serve the requested file with appropriate content type
		ext := filepath.Ext(path)
		switch ext {
		case ".js":
			w.Header().Set("Content-Type", "application/javascript")
		case ".css":
			w.Header().Set("Content-Type", "text/css")
		case ".html":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case ".json":
			w.Header().Set("Content-Type", "application/json")
		case ".png":
			w.Header().Set("Content-Type", "image/png")
		case ".jpg", ".jpeg":
			w.Header().Set("Content-Type", "image/jpeg")
		case ".svg":
			w.Header().Set("Content-Type", "image/svg+xml")
		case ".ico":
			w.Header().Set("Content-Type", "image/x-icon")
		}

		if _, err := w.Write(data); err != nil {
			// Log error but don't return error response as headers are already sent
			// This follows standard HTTP handler patterns
			_ = err // Explicitly ignore error as response cannot be changed
		}
	}, nil
}
