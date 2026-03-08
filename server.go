package main

import (
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/yuin/goldmark"
)

//go:embed templates/*.html
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type app struct {
	vaultRoot string
	vaultName string
	pages     map[string]*template.Template // pre-parsed layout+page pairs
	md        goldmark.Markdown
	noteIndex map[string]string
	fileIndex map[string]string
}

func serve(vaultRoot, addr string) error {
	noteIndex := buildNoteIndex(vaultRoot)
	fileIndex := buildFileIndex(vaultRoot)

	a := &app{
		vaultRoot: vaultRoot,
		vaultName: filepath.Base(vaultRoot),
		md:        newMarkdown(noteIndex, fileIndex),
		noteIndex: noteIndex,
		fileIndex: fileIndex,
	}

	// Pre-parse each page template paired with the layout once at startup.
	// Auto-discover all page templates from the embedded filesystem.
	entries, err := fs.ReadDir(templateFS, "templates")
	if err != nil {
		return fmt.Errorf("reading template directory: %w", err)
	}

	a.pages = make(map[string]*template.Template)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip the layout itself, only parse page templates
		if name == "layout.html" || filepath.Ext(name) != ".html" {
			continue
		}
		t, err := template.ParseFS(templateFS, "templates/layout.html", "templates/"+name)
		if err != nil {
			return fmt.Errorf("parsing template %s: %w", name, err)
		}
		a.pages[name] = t
	}

	mux := http.NewServeMux()

	// Static files
	staticSub, err := fs.Sub(staticFS, "static")
	if err != nil {
		return fmt.Errorf("static fs: %w", err)
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// API
	mux.HandleFunc("GET /api/tree", a.handleTree)

	// Views
	mux.HandleFunc("GET /view/{path...}", a.handleView)
	mux.HandleFunc("GET /raw/{path...}", a.handleRaw)
	mux.HandleFunc("GET /", a.handleIndex)

	return http.ListenAndServe(addr, mux)
}
