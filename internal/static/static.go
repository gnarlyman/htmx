package static

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed css/* js/*
var embeddedFiles embed.FS

// GetFileSystem returns an http.FileSystem that serves the embedded static files
// for a specific directory (e.g., "css", "js")
func GetFileSystem(dir string) http.FileSystem {
	// Get the embedded files for the specified directory
	fsys, err := fs.Sub(embeddedFiles, dir)
	if err != nil {
		panic(err)
	}
	return http.FS(fsys)
}

// GetCSSFileSystem returns an http.FileSystem that serves the embedded CSS files
func GetCSSFileSystem() http.FileSystem {
	return GetFileSystem("css")
}

// GetJSFileSystem returns an http.FileSystem that serves the embedded JS files
func GetJSFileSystem() http.FileSystem {
	return GetFileSystem("js")
}
