# claude-code-share

A simple HTTP server that reads Claude Code conversation logs (JSONL) and serves them as a browsable web UI for team sharing.

## Install

```bash
go install github.com/nhosoya/claude-code-share@latest
```

Or build from source:

```bash
git clone https://github.com/nhosoya/claude-code-share.git
cd claude-code-share
go build -o claude-code-share .
```

## Usage

```bash
# Start with defaults (port 3333, reads ~/.claude/projects)
./claude-code-share

# Custom port and host
./claude-code-share --port 8080 --host 127.0.0.1

# Custom log directory
./claude-code-share --log-dir /path/to/claude/projects
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--port` | `3333` | HTTP server port |
| `--host` | `0.0.0.0` | HTTP server host (LAN-accessible by default) |
| `--log-dir` | `~/.claude/projects` | Path to Claude Code projects directory |

## Pages

- **`/`** — Project list with session counts and last activity
- **`/projects/{slug}`** — Session list for a project
- **`/sessions/{slug}/{sessionId}`** — Full conversation view with collapsible tool calls

## Development

```bash
# Run tests
go test ./...

# Format code
gofmt -w .

# Build
go build -o claude-code-share .
```
