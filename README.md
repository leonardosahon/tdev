# tdev — tmux sessions from YAML

`tdev` is a tiny Go utility that creates and attaches to tmux sessions from a simple YAML file. It bootstraps windows and panes, sets working directories, and injects startup commands so your development environment is ready in one go.

> Note: tdev assumes your tmux indexes start at 1 (not 0).

## Features

- Define sessions declaratively in YAML
- Create windows with per-window working directories and optional startup commands
- Split panes (vertical by default; horizontal when specified) with per-pane working directories and commands
- Smart path handling: supports `~/` expansion and paths relative to a session `root`
- Idempotent attach: if the session already exists, it simply attaches
- Dry run mode to preview all tmux commands

## Requirements

- tmux installed and on your PATH
- Go toolchain to build (or use the provided build script)

Recommended tmux options (required for correct window/pane numbering):

```
set -g base-index 1
set -g pane-base-index 1
```

## Install

Build locally (static, trimmed binary):

```
./build.sh
```

This produces a `tdev` binary in the repo root. Place it on your PATH, e.g.:

```
mv tdev /usr/local/bin/
```

Alternatively, build directly with Go:

```
CGO_ENABLED=0 go build -ldflags="-s -w" -trimpath -buildvcs=false -o tdev
```

## Usage

Basic:

```
tdev path/to/session.yaml
```

Dry run (prints the tmux commands instead of running them):

```
tdev path/to/session.yaml -d
```

Note the flag order: config path first, then `-d`.

If a tmux session with the same name already exists, `tdev` attaches to it. Otherwise it creates the session, windows, and panes as defined, runs any configured commands, selects the first window, and attaches.

## Configuration

A session file describes the desired tmux layout.

Top-level fields:

- `name` (string): tmux session name
- `root` (string): base directory for relative paths; `~/` is expanded
- `windows` (array): list of windows to create in order

Window fields:

- `name` (string): window name
- `path` (string, optional): working directory for the window (relative to `root` or absolute)
- `cmd` (string, optional): command sent to the window after creation
- `panes` (array, optional): when present, panes are created instead of using `cmd`

Pane fields:

- `path` (string): working directory for the pane (relative to `root` or absolute)
- `cmd` (string, optional): command sent to the pane
- `horizontal` (bool, optional): when true, split is horizontal (`tmux -h`); default is vertical (`tmux -v`)

Behavior notes:

- When a window has panes, `tdev` creates the window at the first pane's path, then splits for subsequent panes in order.
- Commands are injected with `tmux send-keys ... C-m`.
- Only `~/` tilde expansion is performed; other shell expansions are not.

## Example

See `example/example-config.yaml` (uses Nerd Font symbols for names). A plain variant:

```yaml
# Session Name
name: my_session
# Working directory
root: ~/apps/my-awesome-app
windows:
  # Home window
  - name: home
    path: .

  # Backend window
  - name: api
    path: apps/api
    cmd: vim .

  # Frontend window
  - name: web
    path: apps/web
    cmd: vim .

  # Live servers (two panes)
  - name: servers
    panes:
      - path: apps/web
        cmd: pnpm dev
      - path: apps/api
        cmd: air
```

## Troubleshooting

- Session does not appear: ensure tmux is installed and you can run `tmux` in the shell.
- Wrong window/pane numbering: set `base-index` and `pane-base-index` to `1` in your tmux config.
- Paths not found: remember that relative paths resolve against `root`, and only `~/` is expanded.
- Commands not running: `cmd` is sent via `send-keys`; confirm your shell and environment initialization within tmux.

## Development

- Module: `github.com/leonardosahon/tdev`
- Go version in `go.mod`: see repo (`go 1.25.1`)
- Dependency: `gopkg.in/yaml.v3`

## Fonts in the example

`example/note.txt` references Nerd Font icons used in the example names. They are optional. If you see garbled symbols, either install a Nerd Font or use plain names.
