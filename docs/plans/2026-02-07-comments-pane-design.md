# Comments Pane Design

## Overview

Add a dedicated comments pane as a third column in the main view. Comments are removed from the detail panel and displayed in their own scrollable, focusable pane — only visible when the selected issue has comments. Comments are sorted newest-first.

## Layout

Three-column layout when comments exist:

```
┌─ Issues ──────┬─ Detail ──────┬─ 💬 Comments (3) ┐
│               │               │                  │
│  issue list   │  issue detail │  💬 Author       │
│               │  (no comments │  comment text    │
│               │   rendered    │                  │
│               │   here)       │  💬 Author       │
│               │               │  comment text    │
└───────────────┴───────────────┴──────────────────┘
```

Two-column layout when no comments:

```
┌─ Issues ──────┬─ Detail ─────────────────────────┐
│               │                                  │
│  issue list   │  issue detail                    │
│               │                                  │
└───────────────┴──────────────────────────────────┘
```

### Width allocation

The comments pane width comes from splitting the detail area. The issue list width is unaffected.

Non-collapsed with comments:
```
listOuter     = width * listRatio
remainingWidth = width - listOuter
detailOuter   = remainingWidth / 2
commentsOuter = remainingWidth - detailOuter   // avoids rounding loss
innerDetail   = detailOuter - 2                // border overhead
innerComments = commentsOuter - 2
```

### Collapsed mode

When list is collapsed (`ctrl+e`) and comments exist: detail + comments side by side, splitting the full width in half. When no comments: detail fullscreen (unchanged).

### Commenting mode

When `a.commenting` is true (`C` key), the comment input replaces everything — same behavior as today. The comments pane is not visible during comment input.

## Comment rendering

Each comment displays as:

```
💬 Author Name (2026-02-07 14:30)
Comment text here...

💬 Author Name (2026-02-06 10:15)
Older comment text here...
```

Sorted **newest first** (reverse the `issue.Comments` slice). A 💬 icon before each author name provides a visual anchor for where each comment starts.

### Panel title

The title includes the comment count: `💬 Comments (5)`.

## Focus & navigation

The comments pane is a `viewport.Model` — scrollable independently. Tab cycles through available panes:

- **With comments:** list → detail → comments → list
- **Without comments:** list → detail → list (unchanged)

A new `commentsPane` value is added to the `pane` enum. The focused pane gets the highlight border, same as today.

### Focus transitions when comments appear/disappear

| Current Focus  | Transition              | Action                            |
|----------------|-------------------------|-----------------------------------|
| `commentsPane` | New issue has no comments | Move focus to `detailPane`        |
| `commentsPane` | Issue deleted            | Move focus to `detailPane`, clear viewport |
| `listPane`     | New issue has comments   | No change                         |
| `detailPane`   | New issue has comments   | No change                         |

## Implementation

### Files to change

1. **`app.go`**
   - Add `commentsPane` to `pane` enum
   - Add `comments viewport.Model` field to `App` struct
   - Initialize the viewport in `NewApp` with `viewport.New(0, 0)`
   - Add `case commentsPane:` to the focus-based panel routing switch in `Update()`
   - In `issueDetailLoadedMsg` handler: populate comments viewport with `renderComments()`, clear it when no comments, call `resizePanels()`, fall back focus from `commentsPane` to `detailPane` if no comments
   - In `issueDeletedMsg` handler: clear comments viewport, move focus from `commentsPane` to `detailPane`

2. **`view.go`**
   - Update `View()` to render the third column when `a.selected` has comments (both collapsed and non-collapsed cases)
   - Update `resizePanels()` with comments-aware width calculation (split detail area when comments exist, size both viewports)

3. **`keyhandling.go`**
   - Update Tab cycling: list → detail → comments (if comments) → list

4. **`detail.go`**
   - Remove inline comments section (lines 48-61)
   - Add `renderComments()` function that iterates comments in reverse order, renders each with 💬 icon, author, timestamp, and text

5. **`styles.go`**
   - Add `iconComment` constant
   - Update `modeHints()` to handle `commentsPane` focus (use same hints as `detailPane`, or a comments-specific set)

6. **`help.go`**
   - Update tab hint text to reflect three-pane cycling

No new files. No changes to model, API, service, or statusbar layers. Comments data comes from the existing `issueDetailLoadedMsg.issue.Comments` — no new fetching needed.
