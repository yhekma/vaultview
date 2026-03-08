package main

import (
	"embed"
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
	templates *template.Template
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

	tmpl, err := template.ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return err
	}
	a.templates = tmpl

	mux := http.NewServeMux()

	// Static files
	staticSub, _ := fs.Sub(staticFS, "static")
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticSub))))

	// API
	mux.HandleFunc("GET /api/tree", a.handleTree)

	// Views
	mux.HandleFunc("GET /view/{path...}", a.handleView)
	mux.HandleFunc("GET /raw/{path...}", a.handleRaw)
	mux.HandleFunc("GET /", a.handleIndex)

	return http.ListenAndServe(addr, mux)
}
