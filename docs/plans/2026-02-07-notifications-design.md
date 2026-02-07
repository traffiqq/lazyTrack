# Notifications / Mentions Feature Design

## Overview

Add a notifications system that shows issues and comments where the current user was mentioned. Includes an unread count indicator in the status bar and an overlay dialog for browsing and navigating to mentions.

## Data Source

Uses YouTrack's Activities API (`/api/activitiesPage`) to fetch mention-related activities with precise timestamps and context. This gives granular per-event data (who mentioned you, where, when) rather than just a list of issues.

## Data Model & API Layer

### New Model Type

`model.Activity`:
- `ID` (string) — activity ID
- `Timestamp` (int64) — unix milliseconds
- `Author` (*User) — who performed the action
- `TargetID` (string) — issue readable ID (e.g., "PROJ-123")
- `TargetSummary` (string) — issue summary
- `Category` (string) — e.g., "CommentsCategory", "DescriptionCategory"
- `Text` (string) — comment excerpt or description snippet containing the mention

### New API Method

`Client.ListMentionActivities(userLogin string, after int64) ([]Activity, error)`
- Calls `GET /api/activitiesPage`
- Parameters: `categories=CommentsCategory,DescriptionCategory`, `$top=50`, `fields=...`
- Filters response for activities where content mentions `@userLogin`
- `after` parameter: unix-millis timestamp, only fetch activities newer than this

### IssueService Interface

Add: `ListMentionActivities(userLogin string, after int64) ([]Activity, error)`

## Notification Dialog Component

New file: `internal/ui/notification_dialog.go`

Follows the same pattern as `FinderDialog`.

### Structure

- `active bool` — whether the dialog is open
- `notifications []model.Activity` — fetched mention activities
- `cursor int` — highlighted item index
- `loading bool` — true while fetching
- `err string` — error message if fetch fails
- `viewport viewport.Model` — scrolling for long lists

### Lifecycle

- **Open()** — sets `active = true`, returns `tea.Cmd` that fetches mentions (passes current user login and `lastCheckedMentions` timestamp)
- **Close()** — sets `active = false`, updates `lastCheckedMentions` to current time in app state
- **Select (Enter)** — closes dialog, navigates to target issue (collapse list, focus detail, fetch issue by readable ID — same as finder)

### Key Bindings (when active)

- `j/k` or arrows — move cursor
- `Enter` — jump to selected issue, mark as read
- `Esc` — close dialog

### Rendering

- Centered overlay box (same styling as finder)
- Title: "Mentions"
- Each row: `[NEW]` badge (if timestamp > lastChecked) + `ISSUE-ID · Author · time-ago · "comment excerpt..."`
- Selected row highlighted with cursor style

## Integration

### App State Changes (`app.go`)

New fields:
- `notifDialog NotificationDialog`
- `currentUser *model.User` — fetched once at startup via `GetCurrentUser()`
- `lastCheckedMentions int64` — loaded from and saved to state file
- `unreadMentionCount int` — cached count for status bar display

### State Persistence (`config/state.go`)

Add to `UIState`:
- `LastCheckedMentions int64 \`yaml:"last_checked_mentions"\``

Saved on quit alongside other state. Also updated when the notification dialog is closed.

### Init Changes

`Init()` fetches current user in parallel with issues. After user is loaded, a background fetch of mentions begins.

### Key Handling (`keyhandling.go`)

- When `notifDialog.active`: route all keys to dialog's `Update()`. On submit, navigate to issue.
- New global key `n`: opens the notification dialog

### Status Bar (`statusbar.go`)

After the app name, if `unreadMentionCount > 0`, render a badge: `· 3 mentions` with a highlight color (e.g., yellow or the accent color).

### Message Types (`messages.go`)

- `currentUserLoadedMsg{user *model.User}`
- `mentionsLoadedMsg{activities []model.Activity}`

## Flow

1. App starts → fetches current user + issues in parallel
2. After current user loaded → background fetch of mention activities, count unread (newer than `lastCheckedMentions`)
3. Status bar shows unread count badge
4. User presses `n` → opens notification dialog, shows mentions
5. User navigates with j/k, presses Enter → dialog closes, navigates to issue, `lastCheckedMentions` updates to now
6. User presses Esc → dialog closes, `lastCheckedMentions` updates to now
7. Pressing `r` (refresh) also re-fetches mentions and updates the count

## Files to Create/Modify

**New files:**
- `internal/model/activity.go` — Activity type
- `internal/api/activities.go` — ListMentionActivities API method
- `internal/api/activities_test.go` — tests
- `internal/ui/notification_dialog.go` — dialog component

**Modified files:**
- `internal/ui/service.go` — add ListMentionActivities to interface
- `internal/ui/app.go` — new fields, init changes
- `internal/ui/keyhandling.go` — dialog routing + `n` key binding
- `internal/ui/statusbar.go` — unread count badge
- `internal/ui/messages.go` — new message types
- `internal/config/state.go` — LastCheckedMentions field
- `internal/config/state_test.go` — test updates
