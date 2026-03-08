package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// calloutTransformer transforms blockquotes that start with [!type] into callouts.
type calloutTransformer struct{}

var calloutRegex = regexp.MustCompile(`^\[!(\w+)\]\s*(.*)$`)

func (t *calloutTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	source := reader.Source()
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		bq, ok := n.(*ast.Blockquote)
		if !ok {
			return ast.WalkContinue, nil
		}

		firstText := extractFirstText(bq, source)
		if firstText == "" {
			return ast.WalkContinue, nil
		}

		matches := calloutRegex.FindStringSubmatch(firstText)
		if matches == nil {
			return ast.WalkContinue, nil
		}

		calloutType := strings.ToLower(matches[1])
		title := matches[2]
		if title == "" {
			// Capitalize the type as default title
			title = strings.ToUpper(calloutType[:1]) + calloutType[1:]
		}

		bq.SetAttributeString("data-callout", []byte(calloutType))
		bq.SetAttributeString("data-callout-title", []byte(title))

		// Remove the first paragraph (the [!type] line) from children
		if first := bq.FirstChild(); first != nil {
			if _, ok := first.(*ast.Paragraph); ok {
				bq.RemoveChild(bq, first)
			}
		}

		return ast.WalkContinue, nil
	})
}

func extractFirstText(bq *ast.Blockquote, source []byte) string {
	for child := bq.FirstChild(); child != nil; child = child.NextSibling() {
		if para, ok := child.(*ast.Paragraph); ok {
			var sb strings.Builder
			for inline := para.FirstChild(); inline != nil; inline = inline.NextSibling() {
				if t, ok := inline.(*ast.Text); ok {
					sb.Write(t.Segment.Value(source))
				}
			}
			return sb.String()
		}
	}
	return ""
}

// calloutRenderer replaces the default blockquote renderer to add callout styling.
type calloutRenderer struct{}

func (r *calloutRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindBlockquote, r.renderBlockquote)
}

var calloutIcons = map[string]string{
	"note":      "📝",
	"abstract":  "📋",
	"summary":   "📋",
	"info":      "ℹ️",
	"todo":      "✅",
	"tip":       "💡",
	"hint":      "💡",
	"important": "❗",
	"success":   "✅",
	"check":     "✅",
	"done":      "✅",
	"question":  "❓",
	"help":      "❓",
	"faq":       "❓",
	"warning":   "⚠️",
	"caution":   "⚠️",
	"attention": "⚠️",
	"failure":   "❌",
	"fail":      "❌",
	"missing":   "❌",
	"danger":    "🔴",
	"error":     "🔴",
	"bug":       "🐛",
	"example":   "📖",
	"quote":     "💬",
	"cite":      "💬",
}

func (r *calloutRenderer) renderBlockquote(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	bq := node.(*ast.Blockquote)

	calloutTypeAttr, hasCallout := bq.AttributeString("data-callout")
	if !hasCallout {
		if entering {
			w.WriteString("<blockquote>\n")
		} else {
			w.WriteString("</blockquote>\n")
		}
		return ast.WalkContinue, nil
	}

	ctBytes, ok := calloutTypeAttr.([]byte)
	if !ok {
		// Attribute exists but wrong type — fall back to regular blockquote
		if entering {
			w.WriteString("<blockquote>\n")
		} else {
			w.WriteString("</blockquote>\n")
		}
		return ast.WalkContinue, nil
	}
	calloutType := string(ctBytes)
	icon := calloutIcons[calloutType]
	if icon == "" {
		icon = "📝"
	}

	if entering {
		titleAttr, _ := bq.AttributeString("data-callout-title")
		title := ""
		if tb, ok := titleAttr.([]byte); ok {
			title = string(tb)
		}
		fmt.Fprintf(w, "<div class=\"callout callout-%s\">\n", calloutType)
		fmt.Fprintf(w, "<div class=\"callout-title\">%s %s</div>\n", icon, title)
		w.WriteString("<div class=\"callout-content\">\n")
	} else {
		w.WriteString("</div>\n</div>\n")
	}
	return ast.WalkContinue, nil
}

// --- Extension ---

type calloutExtension struct{}

func (e *calloutExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithASTTransformers(
			util.Prioritized(&calloutTransformer{}, 99),
		),
	)
	m.Renderer().AddOptions(
		renderer.WithNodeRenderers(
			util.Prioritized(&calloutRenderer{}, 99),
		),
	)
}
