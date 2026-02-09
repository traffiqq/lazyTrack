# lazyTrack

A terminal UI for managing YouTrack issues, inspired by lazygit. Built with Go and the [Charm](https://charm.sh/) stack (Bubble Tea, Lip Gloss, Bubbles).

## Features

- **Issue CRUD** — create, view, edit, and delete issues
- **Vim editor integration** — edit issues in your `$EDITOR` (nvim/vim/vi) with YAML front matter
- **Comments** — view inline comments and add new ones from the TUI
- **State & assignee management** — set issue state and assign users with autocomplete search
- **Leader key** — `space` opens a which-key style popup showing all available actions
- **Fuzzy finder** — search issues by title with type filters (Bug/Task)
- **YouTrack query search** — filter issues using YouTrack query syntax (e.g., `project: PROJ #Unresolved sort by: updated`)
- **Quick filters** — toggle "Assigned to me", "Bug", and "Task" filters with `1`/`2`/`3`
- **Project picker** — switch project context on the fly
- **Go to issue** — jump directly to an issue by number
- **Mention tracking** — view issues mentioning you, with unread count in the status bar
- **Resizable multi-panel layout** — issue list, detail, and comments panels with adjustable split and collapsible list
- **Session persistence** — restores selected issue, panel ratio, and collapse state across sessions
- **First-run setup wizard** — guided configuration on first launch

## AI-Assisted Development

This project was developed by a human engineer with AI assistance for implementation, code generation, and iteration. Architecture decisions, feature design, and code review were human-driven.

## Installation

### Homebrew

```bash
brew install traffiqq/tap/lazytrack
```

### GitHub Releases

Download the binary for your platform from [GitHub Releases](https://github.com/traffiqq/lazytrack/releases). Available for Linux and macOS (amd64/arm64).

### From source

```bash
git clone https://github.com/traffiqq/lazytrack.git
cd lazytrack
make build
```

## Quick Start

1. Run `lazytrack`
2. On first launch, the setup wizard will prompt for your YouTrack server URL and a permanent token
3. The main view has two panels: **issue list** (left) and **issue detail** (right)
4. Press `space` to see available actions, or `?` for the full keybinding reference

## Requirements

- A terminal with a [Nerd Font](https://www.nerdfonts.com/) installed for panel title icons. Without a Nerd Font, icons render as placeholder characters but the app remains fully functional.

## Configuration

Config is stored at `~/.config/lazytrack/config.yaml` (XDG-compliant, created automatically by the setup wizard):

```yaml
server:
  url: "https://youtrack.example.com"
  token: "perm:your-permanent-token-here"
```

## Keybindings

### Navigation

| Key | Action |
|---|---|
| `j`/`k`, `up`/`down` | Navigate issues / scroll detail |
| `tab` | Cycle panels (list → detail → comments) |
| `enter` | Load issue detail |
| `#` | Go to issue by number |
| `r` | Refresh issues, detail, and mentions |

### Quick Filters

| Key | Action |
|---|---|
| `1` | Toggle "Assigned to me" |
| `2` | Toggle "Bug" type filter |
| `3` | Toggle "Task" type filter |

### Leader Key Actions (`space` + key)

| Key | Action |
|---|---|
| `space c` | Create issue |
| `space e` | Edit issue |
| `space d` | Delete issue (confirm with `y`/`n`) |
| `space m` | Add comment |
| `space s` | Set state |
| `space a` | Assign issue |
| `space p` | Select project |
| `space f` | Find issue (fuzzy finder) |
| `space n` | View mentions |
| `space t` | Toggle issue list panel |
| `space v` | Edit issue in vim |

### General

| Key | Action |
|---|---|
| `/` | Search with YouTrack query |
| `H`/`L`, `ctrl+left`/`ctrl+right` | Resize panels |
| `?` | Toggle help |
| `q`, `ctrl+c` | Quit |

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run go vet
```
