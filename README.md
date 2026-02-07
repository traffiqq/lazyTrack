# lazyTrack

A terminal UI for managing YouTrack issues, inspired by lazygit. Built with Go and the [Charm](https://charm.sh/) stack (Bubble Tea, Lip Gloss, Bubbles).

## Features

- **Issue CRUD** — create, view, edit, and delete issues
- **Comments** — add comments directly from the TUI
- **State & assignee management** — set issue state and assign users inline
- **Fuzzy finder** — search issues with type filters (Bug/Task)
- **YouTrack query search** — filter issues using YouTrack query syntax (e.g., `project: PROJ #Unresolved sort by: updated`)
- **Resizable two-panel layout** — issue list on the left, detail on the right, with adjustable split
- **First-run setup wizard** — guided configuration on first launch

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
4. Press `?` to see all keybindings

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
| `j`/`k`, `up`/`down` | Navigate issues |
| `tab` | Switch panels |
| `enter` | Load issue detail |

### Actions

| Key | Action |
|---|---|
| `c` | Create issue |
| `e` | Edit issue |
| `d` | Delete issue (confirm with `y`/`n`) |
| `C` | Add comment (`ctrl+d` to submit, `esc` to cancel) |
| `/` | Search/filter with YouTrack query (`enter` to apply, `esc` to cancel) |
| `f` | Find issue (fuzzy finder) |
| `s` | Set state (`enter` to apply, `esc` to cancel) |
| `a` | Assign issue (`enter` to apply, `esc` to cancel) |

### General

| Key | Action |
|---|---|
| `H`/`L`, `ctrl+left`/`ctrl+right` | Resize panels |
| `ctrl+e` | Toggle issue list |
| `?` | Toggle help |
| `q`, `ctrl+c` | Quit |

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run go vet
```
