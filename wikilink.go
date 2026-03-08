package main

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// --- AST Nodes ---

// WikilinkNode is an inline node for [[wikilinks]].
var KindWikilink = ast.NewNodeKind("Wikilink")

type WikilinkNode struct {
	ast.BaseInline
	Target  string // the note/file name
	Alias   string // display text (from [[target|alias]])
	Heading string // anchor (from [[target#heading]])
	IsEmbed bool   // ![[embed]] vs [[link]]
}

func (n *WikilinkNode) Kind() ast.NodeKind { return KindWikilink }
func (n *WikilinkNode) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{
		"Target": n.Target,
		"Alias":  n.Alias,
	}, nil)
}

// --- Parser ---

type wikilinkParser struct{}

func (p *wikilinkParser) Trigger() []byte {
	return []byte{'[', '!'}
}

func (p *wikilinkParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()
	if len(line) < 4 {
		return nil
	}

	isEmbed := false
	start := 0

	// Check for ![[
	if line[0] == '!' && len(line) > 4 && line[1] == '[' && line[2] == '[' {
		isEmbed = true
		start = 3
	} else if line[0] == '[' && line[1] == '[' {
		start = 2
	} else {
		return nil
	}

	// Find closing ]]
	end := -1
	for i := start; i < len(line)-1; i++ {
		if line[i] == ']' && line[i+1] == ']' {
			end = i
			break
		}
	}
	if end < 0 {
		return nil
	}

	content := string(line[start:end])
	if content == "" {
		return nil
	}

	node := &WikilinkNode{IsEmbed: isEmbed}

	// Parse [[target#heading|alias]]
	if idx := strings.Index(content, "|"); idx >= 0 {
		node.Alias = content[idx+1:]
		content = content[:idx]
	}
	if idx := strings.Index(content, "#"); idx >= 0 {
		node.Heading = content[idx+1:]
		content = content[:idx]
	}
	node.Target = content

	// Advance the reader past the consumed text
	// end + 2 accounts for the content plus ]]
	consumed := end + 2
	if isEmbed {
		consumed++ // account for leading !
	}
	block.Advance(consumed)

	return node
}

// --- Renderer ---

type wikilinkRenderer struct {
	noteIndex map[string]string
	fileIndex map[string]string
}

func (r *wikilinkRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindWikilink, r.renderWikilink)
}

func (r *wikilinkRenderer) renderWikilink(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}

	n := node.(*WikilinkNode)

	if n.IsEmbed {
		return r.renderEmbed(w, n)
	}
	return r.renderLink(w, n)
}

func (r *wikilinkRenderer) renderLink(w util.BufWriter, n *WikilinkNode) (ast.WalkStatus, error) {
	href := r.resolveLink(n.Target, n.Heading)
	display := n.Target
	if n.Alias != "" {
		display = n.Alias
	}
	if n.Heading != "" && n.Alias == "" {
		display = n.Target + " > " + n.Heading
	}

	fmt.Fprintf(w, `<a href="%s" class="wikilink">%s</a>`, href, display)
	return ast.WalkContinue, nil
}

func (r *wikilinkRenderer) renderEmbed(w util.BufWriter, n *WikilinkNode) (ast.WalkStatus, error) {
	target := n.Target

	// Check if it's an image
	ext := strings.ToLower(filepath.Ext(target))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
		src := r.resolveRawPath(target)
		alt := n.Alias
		if alt == "" {
			alt = target
		}
		fmt.Fprintf(w, `<img src="%s" alt="%s" class="embed-image" loading="lazy">`, src, alt)
		return ast.WalkContinue, nil
	case ".mp4", ".webm":
		src := r.resolveRawPath(target)
		fmt.Fprintf(w, `<video src="%s" controls class="embed-video"></video>`, src)
		return ast.WalkContinue, nil
	case ".mp3":
		src := r.resolveRawPath(target)
		fmt.Fprintf(w, `<audio src="%s" controls class="embed-audio"></audio>`, src)
		return ast.WalkContinue, nil
	case ".pdf":
		src := r.resolveRawPath(target)
		fmt.Fprintf(w, `<iframe src="%s" class="embed-pdf" width="100%%" height="600"></iframe>`, src)
		return ast.WalkContinue, nil
	}

	// Note embed — render as a linked blockquote placeholder
	href := r.resolveLink(target, n.Heading)
	display := target
	if n.Alias != "" {
		display = n.Alias
	}
	fmt.Fprintf(w, `<blockquote class="embed-note"><p>📄 <a href="%s">%s</a></p></blockquote>`, href, display)
	return ast.WalkContinue, nil
}

func (r *wikilinkRenderer) resolveLink(target, heading string) string {
	// Look up in note index
	if relPath, ok := r.noteIndex[target]; ok {
		u := "/view/" + url.PathEscape(strings.TrimSuffix(relPath, ".md"))
		if heading != "" {
			u += "#" + url.PathEscape(heading)
		}
		return u
	}
	// Fallback: treat as-is
	u := "/view/" + url.PathEscape(target)
	if heading != "" {
		u += "#" + url.PathEscape(heading)
	}
	return u
}

func (r *wikilinkRenderer) resolveRawPath(target string) string {
	if relPath, ok := r.fileIndex[target]; ok {
		return "/raw/" + relPath
	}
	return "/raw/" + target
}

// --- Extension ---

type wikilinkExtension struct {
	noteIndex map[string]string
	fileIndex map[string]string
}

func (e *wikilinkExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(&wikilinkParser{}, 199),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&wikilinkRenderer{
				noteIndex: e.noteIndex,
				fileIndex: e.fileIndex,
			}, 199),
		),
	)
}
