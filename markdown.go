package main

import (
	"bytes"

	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/mermaid"
)

// RenderResult holds the rendered HTML and extracted metadata.
type RenderResult struct {
	HTML        string
	Frontmatter map[string]interface{}
}

func newMarkdown(noteIndex, fileIndex map[string]string) goldmark.Markdown {
	return goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			meta.Meta,
			highlighting.NewHighlighting(
				highlighting.WithStyle("dracula"),
			),
			&mermaid.Extender{},
			&wikilinkExtension{
				noteIndex: noteIndex,
				fileIndex: fileIndex,
			},
			&calloutExtension{},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)
}

func renderMarkdown(md goldmark.Markdown, source []byte) (*RenderResult, error) {
	ctx := parser.NewContext()
	var buf bytes.Buffer
	if err := md.Convert(source, &buf, parser.WithContext(ctx)); err != nil {
		return nil, err
	}

	fm := meta.Get(ctx)

	return &RenderResult{
		HTML:        buf.String(),
		Frontmatter: fm,
	}, nil
}
