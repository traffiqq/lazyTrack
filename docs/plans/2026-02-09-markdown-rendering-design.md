# Markdown Rendering in Issue Detail & Comments

## Overview

Add basic markdown rendering to issue descriptions and comments using `charmbracelet/glamour` with a custom style matching the lazyTrack color palette.

## Motivation

YouTrack issue descriptions and comments contain markdown. Currently they render as plain text, losing formatting like headings, bold, code blocks, and lists.

## Approach

Use `charmbracelet/glamour` — the Charm stack's terminal markdown renderer. It integrates with Lip Gloss, supports custom `StyleConfig` structs, and handles word wrapping.

### Where it applies

- `renderIssueDetail()` — the `issue.Description` field (detail.go:43)
- `renderComments()` — each `c.Text` field (detail.go:68)

### Fallback

If glamour returns an error for any input, fall back to raw plain text. No crash, no broken UI.

## Custom Style Config

Maps glamour elements to the lazyTrack palette:

| Element | Style | Color | Rationale |
|---|---|---|---|
| Headings (H1-H3) | Bold | Purple `99` | Matches `titleStyle` |
| Bold text | Bold | Default fg | Standard emphasis |
| Italic text | Italic | Default fg | Standard emphasis |
| Code spans (inline) | — | Gray bg `236` | Matches `statusBarStyle` bg |
| Code blocks | Left indent | Gray bg `236` | Same as inline code |
| Links | Underline | Blue `69` | Matches focused accent |
| List bullets | — | Green `78` | Matches active filter color |
| Blockquotes | Left bar | Gray `245` | Matches `hintDescStyle` |
| Horizontal rules | Dashes | Gray `240` | Matches unfocused border |

Skipped elements (YAGNI): tables, images, footnotes, HTML blocks.

No syntax highlighting in code blocks (chroma dependency too large).

## Implementation

### New file: `internal/ui/markdown.go`

- `buildMarkdownStyle() ansi.StyleConfig` — returns custom style config
- `renderMarkdown(text string, width int) string` — creates renderer, calls `Render()`, falls back to plain text on error

### Changes to existing files

**`detail.go`:**
- `renderIssueDetail()` — add `width` parameter, replace raw `issue.Description` with `renderMarkdown(issue.Description, width)`
- `renderComments()` — add `width` parameter, replace raw `c.Text` with `renderMarkdown(c.Text, width)`

**`view.go`:**
- Update call sites of `renderIssueDetail` and `renderComments` to pass `innerWidth` values (already computed in layout branches)

### New file: `internal/ui/markdown_test.go`

- Verify `renderMarkdown()` produces non-empty output for basic markdown
- Verify empty input returns empty string
- No ANSI snapshot tests

### No changes to

Model, API, config, app.go state, key handling. Purely a rendering-layer change.

## Edge Cases

- **Empty string** — short-circuit, return empty
- **Plain text** — glamour passes through with word wrap (improvement over today)
- **glamour error** — return original text unchanged
- **Trailing whitespace** — `strings.TrimRight` the output

## Not in scope

- Toggle between raw/rendered views
- Syntax highlighting in code blocks
- Image rendering
