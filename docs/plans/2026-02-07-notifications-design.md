# Notifications / Mentions Feature Design

## Overview

Add a notifications system that shows issues where the current user was mentioned. Includes an unread count indicator in the status bar and an overlay dialog for browsing and navigating to mentions.

## Data Source

Uses YouTrack's query engine with `mentioned: me` to find issues where the current user is mentioned (in descriptions or comments). This reuses the existing `ListIssues` infrastructure and returns `[]model.Issue` — no new model types needed.

The Activities API (`/api/activitiesPage`) was considered but rejected: it has no server-side mention filter, returns all activities globally, and its polymorphic response structure (`added`/`removed` sub-objects) would require a complex new parsing layer. The query approach is simpler, proven, and sufficient for a first iteration.

## API Layer

### No New API Methods

The existing `ListIssues(query, skip, top)` handles everything. The mention query is constructed as:

```
mentioned: me sort by: updated desc
```

If an active project is set, prepend `project: PROJ` as the existing `effectiveQuery()` pattern does.

### IssueService Interface

No changes needed — `ListIssues` already exists on the interface.

## Notification Dialog Component

New file: `internal/ui/notification_dialog.go`

Follows the same structural pattern as `FinderDialog`.

### Structure

```go
type NotificationDialog struct {
    active        bool
    submitted     bool              // true when user pressed Enter on an item
    selectedIssue *model.Issue      // the issue the user selected
    issues        []model.Issue     // fetched mention results
    cursor        int               // highlighted item index
    loading       bool              // true while fetching
    err           string            // error message if fetch fails
    lastChecked   int64             // timestamp passed in from app state
}
```

Key points vs. the original design:
- `submitted` + `selectedIssue` fields match the FinderDialog pattern for communicating results back to the app via `handleKeyMsg`.
- No `viewport.Model` — uses manual scroll math (start/end index capping) consistent with FinderDialog and ProjectPickerDialog.
- `lastChecked` is passed in on Open so each row can render a `[NEW]` badge.

### Constructor

```go
func NewNotificationDialog() NotificationDialog {
    return NotificationDialog{}
}
```

### Lifecycle

- **Open(lastChecked int64) tea.Cmd** — sets `active = true`, resets `submitted/selectedIssue/cursor/err`, stores `lastChecked`, sets `loading = true`, returns a `tea.Cmd` that fetches mentions via `ListIssues("mentioned: me ...", 0, 50)`.
- **Close()** — sets `active = false`. Does NOT update `lastChecked` — that is the app's responsibility (see Integration section).
- **SetResults(issues []model.Issue)** — stores fetched issues, clears loading state.
- **SetError(errStr string)** — stores error, clears loading state.

### Key Bindings (when active)

- `j/k` or `down/up` — move cursor
- `Enter` — set `submitted = true`, store `selectedIssue`, call `Close()`
- `Esc` — call `Close()`

### Rendering (`View(width, height int) string`)

- Centered overlay box (same styling as FinderDialog: rounded border, color `99`, padding `1,2`)
- Dialog size: 60% width, 70% height (same proportions as finder)
- Title: `"Mentions"`
- Each row: `[NEW]` badge (if issue.Updated > lastChecked) + `ISSUE-ID  Summary`
- Selected row highlighted with background color `237`
- Manual scroll window over results (same math as FinderDialog lines 308-315)
- Bottom hint bar: `j/k: navigate  enter: open  esc: close`
- Loading state: "Loading mentions..."
- Empty state: "No mentions found"
- Error state: styled error message

## Integration

### App State Changes (`app.go`)

New fields on `App`:
- `notifDialog NotificationDialog` — initialized in `NewApp()` via `NewNotificationDialog()`
- `currentUser *model.User` — fetched once at startup via `GetCurrentUser()`
- `lastCheckedMentions int64` — loaded from `state.UI.LastCheckedMentions` in `NewApp()`
- `mentionedIssues []model.Issue` — cached results for status bar count
- `unreadMentionCount int` — count of issues with `Updated > lastCheckedMentions`

### State Persistence (`config/state.go`)

Add to `UIState`:
```go
LastCheckedMentions int64 `yaml:"last_checked_mentions,omitempty"`
```

### Init Changes (`app.go`)

`Init()` returns `tea.Batch(a.fetchIssuesCmd(), a.fetchCurrentUserCmd())`.

New command in `commands.go`:
```go
func (a *App) fetchCurrentUserCmd() tea.Cmd
```
Calls `service.GetCurrentUser()`, returns `currentUserLoadedMsg` on success or `errMsg` on failure.

### Update Handlers (`app.go`)

**`currentUserLoadedMsg` handler:**
- Stores `a.currentUser = msg.user`
- Triggers `a.fetchMentionsCmd()` to load initial mention count

**`mentionsLoadedMsg` handler:**
- Stores `a.mentionedIssues = msg.issues`
- Computes `a.unreadMentionCount` by counting issues where `issue.Updated > a.lastCheckedMentions`
- If `a.notifDialog.active && a.notifDialog.loading`, calls `a.notifDialog.SetResults(msg.issues)`

**`errMsg` handler (existing, extended):**
- Add branch: if `a.notifDialog.active`, call `a.notifDialog.SetError(msg.err.Error())` and return

**`GetCurrentUser()` failure handling:**
- If `GetCurrentUser()` fails, `currentUser` stays nil
- Pressing `n` when `currentUser == nil` sets `a.err = "Could not load user — mentions unavailable"`
- Status bar badge simply doesn't appear (count stays 0)

### Commands (`commands.go`)

New commands:
```go
func (a *App) fetchCurrentUserCmd() tea.Cmd   // -> currentUserLoadedMsg
func (a *App) fetchMentionsCmd() tea.Cmd       // -> mentionsLoadedMsg
```

`fetchMentionsCmd` constructs the query `mentioned: me sort by: updated desc`, prepends active project if set, calls `service.ListIssues(query, 0, 50)`.

### Message Types (`messages.go`)

```go
type currentUserLoadedMsg struct {
    user *model.User
}

type mentionsLoadedMsg struct {
    issues []model.Issue
}
```

### Key Handling (`keyhandling.go`)

**Dialog routing (added before finder check):**
```go
if a.notifDialog.active {
    var cmd tea.Cmd
    a.notifDialog, cmd = a.notifDialog.Update(msg)
    if a.notifDialog.submitted && a.notifDialog.selectedIssue != nil {
        issueID := a.notifDialog.selectedIssue.IDReadable
        a.lastCheckedMentions = latestIssueTimestamp(a.mentionedIssues)
        a.unreadMentionCount = 0
        a.listCollapsed = true
        a.focus = detailPane
        a.resizePanels()
        a.loading = true
        return a, a.fetchDetailCmd(issueID)
    }
    return a, cmd
}
```

**New global key `n`:**
```go
case "n":
    if a.currentUser == nil {
        a.err = "Could not load user — mentions unavailable"
        return a, nil
    }
    a.loading = true
    return a, a.notifDialog.Open(a.lastCheckedMentions)
    // The Open() cmd fetches mentions; result arrives as mentionsLoadedMsg
```

Note: Esc closes the dialog but does NOT update `lastCheckedMentions`. Only selecting an issue (Enter) marks mentions as read. This prevents accidentally marking unseen mentions as read.

### `lastCheckedMentions` Timestamp Strategy

- Set to the **newest issue's `Updated` timestamp** from the fetched results (not `time.Now()`), avoiding clock skew between client and server.
- Only updated when the user **selects an issue** (Enter), not on Esc/dismiss.
- Helper: `latestIssueTimestamp(issues []model.Issue) int64` returns the max `Updated` value from the slice.

### Refresh Integration (`keyhandling.go`)

The `r` key handler adds `a.fetchMentionsCmd()` to the batch:
```go
case "r":
    a.loading = true
    var refreshCmds []tea.Cmd
    refreshCmds = append(refreshCmds, a.fetchIssuesCmd())
    if a.selected != nil {
        issueID := a.selected.IDReadable
        refreshCmds = append(refreshCmds, a.fetchDetailCmd(issueID))
    }
    if a.currentUser != nil {
        refreshCmds = append(refreshCmds, a.fetchMentionsCmd())
    }
    return a, tea.Batch(refreshCmds...)
```

### View Integration (`view.go`)

Add notification dialog to the overlay chain, before the finder:
```go
if a.notifDialog.active {
    return a.notifDialog.View(a.width, a.height)
}
if a.finderDialog.active {
    ...
```

### Status Bar (`statusbar.go`)

After the app name + project/query context, before the loading indicator:
```go
if a.unreadMentionCount > 0 {
    left += mentionBadgeStyle.Render(fmt.Sprintf(" · %d mentions", a.unreadMentionCount))
}
```

The badge uses a highlight color (e.g., yellow `220`). It participates in the existing overflow system — if the terminal is too narrow, hints drop from the right side first; the badge stays on the left but is kept short enough to not cause issues.

### Help Overlay (`help.go`)

Add to the Actions section:
```
  n           Mentions
```

### Quit Handler (`keyhandling.go`)

Add `LastCheckedMentions` to the state saved on quit:
```go
case "ctrl+c", "q":
    state := config.State{
        UI: config.UIState{
            ListRatio:           a.listRatio,
            ListCollapsed:       a.listCollapsed,
            LastCheckedMentions: a.lastCheckedMentions,
        },
    }
    ...
```

## Flow

1. App starts → fetches current user + issues in parallel
2. After current user loaded → fetches mentions (background), computes unread count
3. Status bar shows unread count badge (e.g., `· 3 mentions`)
4. User presses `n` → opens notification dialog, fetches fresh mentions, shows list with `[NEW]` badges
5. User navigates with j/k, presses Enter → dialog closes, navigates to issue, `lastCheckedMentions` updates to newest issue timestamp, unread count resets to 0
6. User presses Esc → dialog closes, `lastCheckedMentions` is NOT updated (unseen mentions stay "new")
7. Pressing `r` (refresh) re-fetches mentions and updates the count
8. On quit, `lastCheckedMentions` is persisted to `state.yaml`

## Files to Create/Modify

**New files:**
- `internal/ui/notification_dialog.go` — NotificationDialog component

**Modified files:**
- `internal/ui/service.go` — no changes needed (ListIssues already exists)
- `internal/ui/app.go` — new fields (`notifDialog`, `currentUser`, `lastCheckedMentions`, `mentionedIssues`, `unreadMentionCount`), `NewApp()` initialization, `Init()` parallel fetch
- `internal/ui/commands.go` — `fetchCurrentUserCmd()`, `fetchMentionsCmd()`, `latestIssueTimestamp()` helper
- `internal/ui/keyhandling.go` — notification dialog routing, `n` key binding, Esc-only-no-update behavior, refresh integration, quit handler update
- `internal/ui/view.go` — add notification dialog to overlay chain
- `internal/ui/statusbar.go` — unread mention count badge
- `internal/ui/styles.go` — `mentionBadgeStyle`
- `internal/ui/messages.go` — `currentUserLoadedMsg`, `mentionsLoadedMsg`
- `internal/ui/help.go` — add `n` keybinding to help text
- `internal/config/state.go` — `LastCheckedMentions` field on `UIState`
- `internal/config/state_test.go` — test round-trip of new field
