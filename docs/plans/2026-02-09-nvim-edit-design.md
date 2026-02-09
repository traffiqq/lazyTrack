# Nvim Edit Design

Open the selected issue in `$EDITOR` (nvim by default) for full editing, similar to `kubectl edit` or `git commit`.

## Trigger

`space+v` leader key ("vim edit"). The existing `space+e` TUI edit dialog remains as an alternative.

## Editor Resolution

Check `$EDITOR` env var, fall back to `nvim`, then `vim`, then `vi`.

## Flow

1. User presses `space+v` with an issue selected.
2. Write issue to temp file `/tmp/lazytrack-edit-<ISSUE_ID>.md` as YAML front matter + markdown body.
3. Suspend TUI via `tea.ExecProcess`, open editor on the temp file.
4. User edits, saves, quits.
5. TUI resumes. Parse temp file, diff against original values.
6. If changed, call `UpdateIssue` with only the changed fields.
7. Clean up temp file. Show success/error in status bar.
8. If nothing changed (user quit without saving), skip API call silently.

## Temp File Format

```markdown
---
summary: Fix login bug
state: In Progress
assignee: johndoe
type: Bug
---

The login page throws a 500 error when the user
enters special characters in the password field.
```

### Writing

- `summary` from `Issue.Summary`
- `state` from `Issue.StateValue()`
- `assignee` from `Issue.AssigneeValue().Login` (empty string if unassigned)
- `type` from `Issue.TypeValue()`
- Body is `Issue.Description`

### Parsing

- Split on `---` delimiters
- Parse front matter as flat `key: value` lines (no YAML library needed)
- Everything after the second `---` (trimmed) is the description
- Malformed file shows error in status bar, no API call

### Diffing

- Compare each field against original values captured before opening editor
- Only include changed fields in `UpdateIssue`
- Empty `assignee:` line means unassign (send null)

## Error Handling

Submit to YouTrack and let it return errors (e.g., invalid state name). Display via existing `errMsg` flow in the status bar.

## Files Changed

- **`internal/ui/editor.go`** (new): `resolveEditor()`, `writeIssueTempFile()`, `parseIssueTempFile()`, `buildEditorUpdateFields()`.
- **`internal/ui/messages.go`**: Add `editorFinishedMsg`.
- **`internal/ui/keyhandling.go`**: Add `case "v":` in leader dispatch.
- **`internal/ui/app.go`**: Handle `editorFinishedMsg` in `Update()`.
- **`internal/ui/styles.go`**: Add `{"v", "vim edit"}` to `leaderHints`.
- **`internal/ui/help.go`**: Add `space v` to help text.

No changes to the API layer.
