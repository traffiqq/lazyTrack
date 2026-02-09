# Markdown Rendering in Issue Detail & Comments

## Overview

Add basic markdown rendering to issue descriptions and comments using `charmbracelet/glamour` with a custom style matching the lazyTrack color palette.

## Motivation

YouTrack issue descriptions and comments contain markdown. Currently they render as plain text, losing formatting like headings, bold, code blocks, and lists.

## Approach

Use `charmbracelet/glamour` — the Charm stack's terminal markdown renderer. It integrates with Lip Gloss, supports custom `StyleConfig` structs, and handles word wrapping.

### Where it applies

Three integration points:

1. `renderIssueDetail()` — `issue.Description` field (detail.go:43), called from `app.go:238` inside `issueDetailLoadedMsg` handler
2. `renderComments()` — each `c.Text` field (detail.go:68), called from `app.go:241` inside `issueDetailLoadedMsg` handler
3. `IssueDialog.renderComments()` — each `c.Text` field (issue_dialog.go:723), called during dialog rendering

### Fallback

If glamour returns an error for any input, fall back to raw plain text. No crash, no broken UI.

## Custom Style Config

Maps glamour elements to the lazyTrack palette:

| Element | Style | Color | Rationale |
|---|---|---|---|
| Document | Margin `0` | — | Prevent glamour's default top/bottom padding |
| Headings (H1-H3) | Bold | Purple `99` | Matches `titleStyle` |
| Bold text | Bold | Default fg | Standard emphasis |
| Italic text | Italic | Default fg | Standard emphasis |
| Code spans (inline) | — | Gray bg `236` | Matches `statusBarStyle` bg |
| Code blocks | Left indent | Gray bg `236` | Same as inline code |
| Links | Underline | Blue `69` | Matches focused accent |
| List bullet prefix | — | Green `78` | Bullet character only via `BlockPrefix` styling |
| List item text | — | Default fg | Keep item text in default color |
| Blockquotes | Left bar | Gray `245` | Matches `hintDescStyle` |
| Horizontal rules | Dashes | Gray `240` | Matches unfocused border |

**Note on list bullets:** glamour's `Item.Color` affects the entire item text, not just the bullet character. To color only the bullet marker green, use a styled `BlockPrefix` on the `Item` field while leaving item text color unset (default fg).

**Note on `StyleConfig` ergonomics:** The `ansi.StylePrimitive` fields `Color` and `BackgroundColor` are `*string` pointers. Use a `stringPtr()` helper for clean assignments.

Skipped elements (YAGNI): tables, images, footnotes, HTML blocks.

### Dependency note

glamour pulls in `yuin/goldmark` (CommonMark parser), `alecthomas/chroma/v2` (syntax highlighting — transitive, unavoidable even though we don't use it), `microcosm-cc/bluemonday` (HTML sanitization), and `muesli/reflow` (text reflow). This roughly doubles the indirect dependency count. Acceptable for the rendering quality gained.

No syntax highlighting in code blocks — chroma is pulled in transitively but we don't invoke it.

## Implementation

### New file: `internal/ui/markdown.go`

- `stringPtr(s string) *string` — helper for `ansi.StylePrimitive` pointer fields
- `buildMarkdownStyle() ansi.StyleConfig` — returns custom style config with `Document.Margin` set to `0`
- `renderMarkdown(text string, width int) string` — creates renderer with `glamour.WithWordWrap(width)`, calls `Render()`, trims trailing whitespace, falls back to plain text on error

### Changes to existing files

**`detail.go`:**
- `renderIssueDetail()` — add `width` parameter, replace raw `issue.Description` with `renderMarkdown(issue.Description, width)`
- `renderComments()` — add `width` parameter, replace raw `c.Text` with `renderMarkdown(c.Text, width)`

**`app.go`:**
- In `issueDetailLoadedMsg` handler (line 234-250): move `a.resizePanels()` call **before** `SetContent` calls so widths are computed before rendering. Pass `a.detail.Width` to `renderIssueDetail()` and `a.comments.Width` to `renderComments()`.
- Add a `reRenderContent()` helper method on `App` that checks `a.selected != nil`, then re-calls `renderIssueDetail(a.selected, a.detail.Width)` + `a.detail.SetContent(...)` and (if comments exist) `renderComments(a.selected.Comments, a.comments.Width)` + `a.comments.SetContent(...)`. Uses `a.selected` directly — no duplicate stored state needed since raw text is already on `a.selected.Description` and `a.selected.Comments`.
- Call `a.reRenderContent()` after `resizePanels()` at every call site where viewport widths change with content already loaded:
  - `tea.WindowSizeMsg` handler (app.go:185)
  - Toggle list collapse — space+t (keyhandling.go:247)
  - Resize left — H key (keyhandling.go:477)
  - Resize right — L key (keyhandling.go:468)

**`issue_dialog.go`:**
- In `IssueDialog.renderComments()` (line 723): replace `c.Text` with `renderMarkdown(c.Text, width)` — width is already available as a parameter.

### Re-rendering on resize

glamour hard-wraps output to the specified width. Unlike plain text (which the viewport soft-wraps), hard-wrapped markdown will overflow or be too short after a terminal/panel resize. Solution:

1. Extract a `reRenderContent()` helper that reads raw text from `a.selected` (already stored — no duplicate state needed) and re-renders with current viewport widths
2. Call `reRenderContent()` after `resizePanels()` at all call sites that change viewport widths with content already loaded:
   - `tea.WindowSizeMsg` (terminal resize)
   - space+t (toggle list collapse)
   - H/L keys (resize panel ratio)
3. Call sites that fetch new issues after `resizePanels()` (notification submit, finder submit, goto) do NOT need re-render — the fetch will re-render with fresh widths
4. `IssueDialog.renderComments` already re-renders on every `View()` call, so no extra work needed there

### New file: `internal/ui/markdown_test.go`

- Verify `renderMarkdown()` produces non-empty output for basic markdown (headings, bold, lists)
- Verify empty input returns empty string
- Verify fallback: invalid glamour scenario returns original text
- Verify width is respected (output lines do not exceed specified width)

### No changes to

Model, API, config, key handling. Rendering-layer change with a helper method addition in `app.go`.

## Edge Cases

- **Empty string** — short-circuit, return empty
- **Plain text** — glamour passes through with word wrap (improvement over today)
- **glamour error** — return original text unchanged
- **Trailing whitespace** — `strings.TrimRight` the output
- **Terminal resize** — re-render markdown at new width (see above)
- **Hardcoded separator** — the `────────` line in `renderIssueDetail` (line 40) stays as-is; it separates metadata from the markdown-rendered description and provides a clear visual boundary

## Not in scope

- Toggle between raw/rendered views
- Syntax highlighting in code blocks
- Image rendering
