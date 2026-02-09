<div align="center">

```
██╗      █████╗ ███████╗██╗   ██╗████████╗██████╗  █████╗  ██████╗██╗  ██╗
██║     ██╔══██╗╚══███╔╝╚██╗ ██╔╝╚══██╔══╝██╔══██╗██╔══██╗██╔════╝██║ ██╔╝
██║     ███████║  ███╔╝  ╚████╔╝    ██║   ██████╔╝███████║██║     █████╔╝
██║     ██╔══██║ ███╔╝    ╚██╔╝     ██║   ██╔══██╗██╔══██║██║     ██╔═██╗
███████╗██║  ██║███████╗   ██║      ██║   ██║  ██║██║  ██║╚██████╗██║  ██╗
╚══════╝╚═╝  ╚═╝╚══════╝   ╚═╝      ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝
```

<em>A simple terminal UI for managing YouTrack issues</em>

</div>

<p align="center">
  <a href="https://github.com/traffiqq/lazytrack/releases"><img src="https://img.shields.io/github/downloads/traffiqq/lazytrack/total" alt="GitHub Downloads"></a>
  <a href="https://goreportcard.com/report/github.com/traffiqq/lazytrack"><img src="https://goreportcard.com/badge/github.com/traffiqq/lazytrack" alt="Go Report Card"></a>
  <a href="https://github.com/traffiqq/lazytrack/releases/latest"><img src="https://img.shields.io/github/v/tag/traffiqq/lazytrack?color=blue" alt="GitHub tag"></a>
  <a href="https://github.com/traffiqq/homebrew-tap"><img src="https://img.shields.io/github/v/tag/traffiqq/lazytrack?label=homebrew&color=blue" alt="Homebrew"></a>
</p>

<p align="center">
  <img src="assets/demo/demo.gif" alt="lazyTrack demo" width="700">
</p>

## Elevator Pitch

You use YouTrack for issue tracking. You've got the browser tab open, you're clicking through menus, waiting for pages to load, losing context every time you switch between your code and your issues. You just want to check what's assigned to you, update a status, or add a quick comment — but instead you're three clicks deep in a web UI that insists on loading a dashboard you never asked for.

lazyTrack brings your issues into the terminal where you already live. Navigate with vim keys, update states with a keystroke, search with YouTrack queries, and never leave your workflow.

## Table of Contents

- [Elevator Pitch](#elevator-pitch)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)
  - [Keybindings](#keybindings)
  - [Leader Key Actions](#leader-key-actions-space--key)
- [Configuration](#configuration)
- [Requirements](#requirements)
- [Development](#development)
- [AI-Assisted Development](#ai-assisted-development)

## Features

### Issue Management

Create, view, edit, and delete issues without leaving the terminal. Full CRUD operations with inline detail view.

### Vim Editor Integration

Press `space v` to open the current issue in your `$EDITOR` (nvim/vim/vi). Issue fields are presented as YAML front matter — edit title, description, state, and assignee in one shot.

### Leader Key Menu

Press `space` and a which-key style popup shows every available action. No need to memorize keybindings — just explore.

![leader_key](assets/demo/leader_key.gif)

### YouTrack Query Search

Press `/` to search using native YouTrack query syntax. Filter by project, state, assignee, or any field YouTrack supports.

![query_search](assets/demo/query_search.gif)

### Quick Filters

Toggle common filters instantly: `1` for "Assigned to me", `2` for bugs, `3` for tasks. Stack them to narrow down fast.

![quick_filters](assets/demo/quick_filters.gif)

### Fuzzy Finder

Press `space f` to fuzzy-search issues by title with type filters. Find what you need in seconds.

### Mention Tracking

See issues mentioning you with an unread count in the status bar. Press `space n` to view them.

### Resizable Multi-Panel Layout

Issue list, detail, and comments panels with adjustable split ratio. Collapse the list panel to focus on the detail view. Layout state persists across sessions.

### Project Picker

Switch between YouTrack projects on the fly with `space p`. Context stays within the selected project.

### Session Persistence

Your selected issue, panel ratio, collapse state, and project context are restored automatically when you relaunch.

## Installation

### Homebrew

Works on macOS and Linux.

```sh
brew install traffiqq/tap/lazytrack
```

### Binary Releases

Download the binary for your platform from [GitHub Releases](https://github.com/traffiqq/lazytrack/releases). Available for Linux and macOS (amd64/arm64).

### From Source

You'll need [Go](https://golang.org/doc/install) installed.

```sh
git clone https://github.com/traffiqq/lazytrack.git
cd lazytrack
make build
```

## Quick Start

1. Run `lazytrack` in your terminal
2. On first launch, the setup wizard prompts for your YouTrack server URL and a [permanent token](https://www.jetbrains.com/help/youtrack/devportal/Manage-Permanent-Token.html)
3. The main view shows two panels: **issue list** (left) and **issue detail** (right)
4. Press `space` to see available actions, or `?` for the full keybinding reference

## Usage

```sh
lazytrack
```

### Keybindings

#### Navigation

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate issues / scroll detail |
| `up` / `down` | Navigate issues / scroll detail |
| `tab` | Cycle panels (list > detail > comments) |
| `enter` | Load issue detail |
| `#` | Go to issue by number |
| `r` | Refresh issues, detail, and mentions |

#### Quick Filters

| Key | Action |
|-----|--------|
| `1` | Toggle "Assigned to me" |
| `2` | Toggle "Bug" type filter |
| `3` | Toggle "Task" type filter |

#### Leader Key Actions (`space` + key)

| Key | Action |
|-----|--------|
| `c` | Create issue |
| `e` | Edit issue |
| `d` | Delete issue (confirm with `y`/`n`) |
| `m` | Add comment |
| `s` | Set state |
| `a` | Assign issue |
| `p` | Select project |
| `f` | Find issue (fuzzy finder) |
| `n` | View mentions |
| `t` | Toggle issue list panel |
| `v` | Edit issue in vim |

#### General

| Key | Action |
|-----|--------|
| `/` | Search with YouTrack query |
| `H` / `L` | Resize panels |
| `ctrl+left` / `ctrl+right` | Resize panels |
| `?` | Toggle help |
| `q` / `ctrl+c` | Quit |

## Configuration

Config is stored at `~/.config/lazytrack/config.yaml` (XDG-compliant). The setup wizard creates it automatically on first run.

```yaml
server:
  url: "https://youtrack.example.com"
  token: "perm:your-permanent-token-here"
```

## Requirements

- A [YouTrack](https://www.jetbrains.com/youtrack/) instance with a permanent token
- A terminal with a [Nerd Font](https://www.nerdfonts.com/) installed for panel title icons (optional — the app is fully functional without one)

## Development

```sh
make build    # Build binary
make test     # Run tests
make lint     # Run go vet
make clean    # Remove binary
```

Built with the [Charm](https://charm.sh/) stack: [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), and [Bubbles](https://github.com/charmbracelet/bubbles).

## AI-Assisted Development

This project was developed by a human engineer with AI assistance for implementation, code generation, and iteration. Architecture decisions, feature design, and code review were human-driven.
