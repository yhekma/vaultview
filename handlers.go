package main

import (
	"encoding/json"
	"html/template"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (a *app) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	// Count notes
	noteCount := len(a.noteIndex)

	data := map[string]interface{}{
		"Title":     "Home",
		"VaultName": a.vaultName,
		"NoteCount": noteCount,
	}

	a.render(w, "landing.html", data)
}

func (a *app) handleView(w http.ResponseWriter, r *http.Request) {
	notePath := r.PathValue("path")
	if notePath == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Security: prevent path traversal
	clean := filepath.Clean(notePath)
	if strings.Contains(clean, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	// Add .md extension if not present
	mdPath := clean
	if !strings.HasSuffix(mdPath, ".md") {
		mdPath += ".md"
	}

	fullPath := filepath.Join(a.vaultRoot, mdPath)

	// Verify the resolved path is within the vault
	resolved, err := filepath.Abs(fullPath)
	if err != nil || !strings.HasPrefix(resolved, a.vaultRoot) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	source, err := os.ReadFile(fullPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	result, err := renderMarkdown(a.md, source)
	if err != nil {
		http.Error(w, "failed to render markdown", http.StatusInternalServerError)
		log.Printf("markdown render error for %s: %v", notePath, err)
		return
	}

	// Derive title from filename
	title := strings.TrimSuffix(filepath.Base(mdPath), ".md")

	data := map[string]interface{}{
		"Title":       title,
		"VaultName":   a.vaultName,
		"Content":     template.HTML(result.HTML),
		"Frontmatter": result.Frontmatter,
	}

	a.render(w, "note.html", data)
}

func (a *app) handleRaw(w http.ResponseWriter, r *http.Request) {
	filePath := r.PathValue("path")
	if filePath == "" {
		http.NotFound(w, r)
		return
	}

	clean := filepath.Clean(filePath)
	if strings.Contains(clean, "..") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	fullPath := filepath.Join(a.vaultRoot, clean)

	resolved, err := filepath.Abs(fullPath)
	if err != nil || !strings.HasPrefix(resolved, a.vaultRoot) {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	// Set content type based on extension
	ext := filepath.Ext(clean)
	ct := mime.TypeByExtension(ext)
	if ct != "" {
		w.Header().Set("Content-Type", ct)
	}

	http.ServeFile(w, r, fullPath)
}

func (a *app) handleTree(w http.ResponseWriter, r *http.Request) {
	tree, err := buildTree(a.vaultRoot)
	if err != nil {
		http.Error(w, "failed to scan vault", http.StatusInternalServerError)
		log.Printf("tree scan error: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
}

func (a *app) render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := a.pages[name]
	if !ok {
		http.Error(w, "unknown page", http.StatusInternalServerError)
		log.Printf("no pre-parsed template for %s", name)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("template execute error: %v", err)
	}
}
