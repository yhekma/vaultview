# VaultView — Project Guide

## Build & Test

```sh
go build ./...          # build — no test suite yet
```

No linter or formatter configured. No CI pipeline. No pre-commit hooks.

## Project Structure

```
main.go          — CLI entry point, flag parsing
server.go        — HTTP server setup, template loading, route registration
handlers.go      — Request handlers: index, view (md + canvas), raw, tree API
markdown.go      — Goldmark markdown pipeline with extensions
wikilink.go      — Goldmark extension: [[wikilink]] parser + renderer
callout.go       — Goldmark extension: > [!type] callout transformer + renderer
tree.go          — Vault directory scanner, note/file index builders
templates/       — Go html/templates (layout.html + page templates, auto-discovered)
static/style.css — All styling, uses CSS custom properties for dark/light theming
```

## Key Patterns

- **Template auto-discovery**: `server.go` reads all `.html` files from `templates/` at startup and pairs each with `layout.html`. Adding a new page template requires no code changes to server.go.
- **Goldmark extensions**: Wikilinks and callouts are implemented as full goldmark extensions (parser + AST node + renderer). Follow the same pattern for new syntax.
- **CSS theming**: All colors use CSS custom properties defined in `:root` (light) and `[data-theme="dark"]` (dark) in `style.css`. Templates use semantic CSS classes (e.g. `.sidebar`, `.tree-file`) instead of Tailwind color classes. Tailwind is only used for layout utilities.
- **Canvas rendering**: `.canvas` files are parsed as JSON server-side, emitted via `template.JS` to avoid HTML escaping, and rendered client-side with positioned divs + SVG edges. URL protocols are validated against an allowlist. Canvas files are indexed in `noteIndex` with the `.canvas` extension as key (matching Obsidian's `[[Foo.canvas]]` wikilink convention).
- **Note index conventions**: `.md` files are keyed without extension (e.g. `"My Note"`), `.canvas` files are keyed with extension (e.g. `"My Canvas.canvas"`). This matches how Obsidian resolves wikilinks. When both `foo.md` and `foo.canvas` exist, `.md` takes priority for extensionless URLs.
- **Security**: User-controlled content (filenames, URLs) must use `textContent` or DOM element creation — never `innerHTML`. External URLs require protocol validation (`http:`, `https:`, `mailto:` only).
- **Embedded assets**: Templates and static files use `go:embed`. The binary is self-contained.

## Conventions

- No test suite exists yet — verify changes compile with `go build ./...`
- Commit messages: imperative mood, first line summarizes the change, body explains why
- Go module path: `vaultview` (not fully qualified)
