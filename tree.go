package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// TreeNode represents a file or directory in the vault.
type TreeNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`               // relative to vault root
	IsDir    bool        `json:"isDir"`
	Children []*TreeNode `json:"children,omitempty"`
}

var excludeDirs = map[string]bool{
	".obsidian": true,
	".git":      true,
	".trash":    true,
	".DS_Store": true,
	"node_modules": true,
}

var includedExts = map[string]bool{
	".md":   true,
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".gif":  true,
	".svg":  true,
	".webp": true,
	".pdf":  true,
	".mp3":  true,
	".mp4":  true,
	".webm": true,
	".csv":    true,
	".canvas": true,
}

// buildTree scans the vault directory and returns a tree of nodes.
func buildTree(vaultRoot string) ([]*TreeNode, error) {
	return scanDir(vaultRoot, vaultRoot)
}

func scanDir(dir, vaultRoot string) ([]*TreeNode, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var nodes []*TreeNode
	for _, e := range entries {
		name := e.Name()

		if strings.HasPrefix(name, ".") && excludeDirs[name] {
			continue
		}
		if excludeDirs[name] {
			continue
		}

		fullPath := filepath.Join(dir, name)
		relPath, _ := filepath.Rel(vaultRoot, fullPath)

		if e.IsDir() {
			children, err := scanDir(fullPath, vaultRoot)
			if err != nil {
				continue
			}
			if len(children) == 0 {
				continue // skip empty dirs
			}
			nodes = append(nodes, &TreeNode{
				Name:     name,
				Path:     relPath,
				IsDir:    true,
				Children: children,
			})
		} else {
			ext := strings.ToLower(filepath.Ext(name))
			if !includedExts[ext] {
				continue
			}
			nodes = append(nodes, &TreeNode{
				Name: name,
				Path: relPath,
				IsDir: false,
			})
		}
	}

	// Sort: directories first, then alphabetical
	slices.SortFunc(nodes, func(a, b *TreeNode) int {
		if a.IsDir != b.IsDir {
			if a.IsDir {
				return -1
			}
			return 1
		}
		return strings.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
	})

	return nodes, nil
}

// buildNoteIndex creates a map from note name to relative path for resolving
// wikilinks. Markdown files are keyed without .md extension (Obsidian convention).
// Canvas files are keyed with .canvas extension (e.g. "My Canvas.canvas").
// If multiple notes share a name, shortest path wins.
func buildNoteIndex(vaultRoot string) map[string]string {
	index := make(map[string]string)
	filepath.Walk(vaultRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if excludeDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		var name string
		switch ext {
		case ".md":
			name = strings.TrimSuffix(info.Name(), ".md")
		case ".canvas":
			name = info.Name() // keep .canvas extension as key
		default:
			return nil
		}
		rel, _ := filepath.Rel(vaultRoot, path)
		if existing, ok := index[name]; !ok || len(rel) < len(existing) {
			index[name] = rel
		}
		return nil
	})
	return index
}

// buildFileIndex creates a map from filename to relative path for resolving
// image/attachment embeds.
func buildFileIndex(vaultRoot string) map[string]string {
	index := make(map[string]string)
	filepath.Walk(vaultRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if excludeDirs[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(vaultRoot, path)
		name := info.Name()
		if _, ok := index[name]; !ok {
			index[name] = rel
		}
		return nil
	})
	return index
}
