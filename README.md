# VaultView

A lightweight, read-only web viewer for [Obsidian](https://obsidian.md) vaults. Point it at a local vault directory and browse your notes in a clean web interface.

Single binary. No build pipeline. No config files.

```
vaultview /path/to/vault
```

## Features

- **File tree sidebar** with collapsible folders and instant filter
- **Markdown rendering** via [goldmark](https://github.com/yuin/goldmark) with GFM (tables, task lists, strikethrough, autolinks)
- **Canvas rendering** — `.canvas` files displayed as interactive pan/zoom diagrams with nodes and edges
- **Dark mode** — toggle between light and dark themes, persisted in browser localStorage
- **Wikilinks** — `[[note]]`, `[[note|alias]]`, `[[note#heading]]` with shortest-path resolution
- **Embeds** — `![[image.png]]`, `![[video.mp4]]`, `![[note]]` rendered inline
- **Callouts** — all Obsidian callout types (`> [!tip]`, `> [!warning]`, etc.) with styled rendering
- **Syntax highlighting** for code blocks (Dracula theme, powered by [chroma](https://github.com/alecthomas/chroma))
- **Mermaid diagrams** rendered client-side (theme-aware)
- **Frontmatter** displayed as a collapsible properties panel
- **All assets embedded** in the binary via `go:embed` for zero-dependency deployment

## Install

Requires Go 1.22+.

```sh
git clone https://github.com/yhekma/vaultview.git
cd vaultview
go build -o vaultview .
```

## Usage

```
Usage: vaultview [flags] <vault-path>

Flags:
  -addr string
        listen address (default "localhost:8080")
```

Examples:

```sh
# Serve on default port (localhost only)
vaultview ~/Documents/MyVault

# Serve on a custom port
vaultview -addr localhost:3000 ~/Documents/MyVault

# Bind to all interfaces (use with caution — see Security below)
vaultview -addr :8080 ~/Documents/MyVault
```

Then open `http://localhost:8080` in your browser.

**Security:** By default, VaultView binds to `localhost` only and is not reachable from other machines. If you use `-addr :PORT` (without `localhost`), the server binds to all network interfaces, exposing your vault to anyone on the network. There is no authentication. Do not expose VaultView on untrusted networks.

## What gets rendered

| Obsidian syntax | Support |
|---|---|
| `[[wikilinks]]` | Links with alias, heading anchor |
| `![[embeds]]` | Images, video, audio, PDF, note placeholders |
| `> [!callout]` | All built-in types with icons and colors |
| `` ```lang `` | Syntax-highlighted code blocks |
| `` ```mermaid `` | Diagrams via mermaid.js CDN |
| YAML frontmatter | Collapsible properties table |
| GFM tables | Standard table rendering |
| `- [x] tasks` | Checkbox task lists |
| `~~strikethrough~~` | Strikethrough text |
| `.canvas` files | Interactive pan/zoom canvas viewer |

## Architecture

```
Browser
  +-- Sidebar (file tree from /api/tree)
  +-- Content pane (rendered markdown or canvas from /view/{path})
  +-- Theme toggle (dark/light, persisted in localStorage)

Go server (net/http)
  GET /              landing page
  GET /view/{path}   render note (.md) or canvas (.canvas) as HTML
  GET /raw/{path}    serve images/attachments
  GET /api/tree      JSON file tree
  GET /static/       embedded CSS
```

Styling uses CSS custom properties for theming with Tailwind CSS utility classes via CDN. Templates and static assets are embedded in the binary.

## Limitations

- **Read-only** — no editing, no saving
- **No search** — filter in the sidebar works on filenames only
- **No plugins** — Dataview, Templater, and other plugin output is not rendered
- **Local use** — no authentication; not intended for public-facing deployment
