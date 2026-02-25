# claude-code-share

A simple HTTP server that reads Claude Code conversation logs (JSONL) and serves them as a browsable web UI for team sharing.

## Tech Stack

- **Go** (1.24+): Single binary, no external dependencies
- **html/template**: Server-side rendering
- **net/http**: Standard library HTTP server
- No JavaScript frameworks — vanilla HTML/CSS only

## Project Structure

```
.
├── main.go              # Entry point, CLI flags, server startup
├── internal/
│   ├── logparser/       # JSONL parsing and data models
│   │   ├── parser.go    # Read/parse JSONL files
│   │   └── models.go    # Data structures for sessions/messages
│   ├── server/          # HTTP handlers and routing
│   │   ├── server.go    # Server setup and routes
│   │   └── handlers.go  # Request handlers
│   └── templates/       # Go embed HTML templates
│       ├── layout.html
│       ├── index.html   # Project list
│       ├── project.html # Session list for a project
│       └── session.html # Conversation view
├── CLAUDE.md
├── PROMPT.md
├── go.mod
└── README.md
```

## JSONL Log Format Reference

Claude Code stores logs at `~/.claude/projects/<project-slug>/<session-uuid>.jsonl`.

Project slug is the workspace path with `/` and `.` replaced by `-`, prefixed with `-`.
Example: `/Users/foo/workspace/my-project` → `-Users-foo-workspace-my-project`

Each line is a JSON object with a `type` field:

### `type: "user"`
```json
{
  "type": "user",
  "uuid": "...",
  "parentUuid": "..." | null,
  "timestamp": "2026-02-25T06:41:55.945Z",
  "sessionId": "d3cd06b0-...",
  "version": "2.1.56",
  "cwd": "/Users/foo/workspace/my-project",
  "message": {
    "role": "user",
    "content": "User's text input or array of tool_result objects"
  }
}
```

Note: `message.content` can be either a string (user text) or an array (tool results from the system).

### `type: "assistant"`
```json
{
  "type": "assistant",
  "uuid": "...",
  "parentUuid": "...",
  "timestamp": "2026-02-25T06:42:02.218Z",
  "sessionId": "d3cd06b0-...",
  "message": {
    "model": "claude-opus-4-6",
    "role": "assistant",
    "content": [
      { "type": "text", "text": "Response text..." },
      { "type": "tool_use", "id": "toolu_01...", "name": "Bash", "input": { "command": "ls", "description": "List files" } }
    ],
    "usage": {
      "input_tokens": 3,
      "output_tokens": 11,
      "cache_creation_input_tokens": 16704,
      "cache_read_input_tokens": 0
    }
  }
}
```

### `type: "progress"`
Subagent progress updates. Contains nested message data.

### `type: "file-history-snapshot"`
File backup tracking. Can be skipped in the viewer.

## Coding Conventions

- Keep it simple — this is a small CLI tool, not a framework
- Use `log/slog` for structured logging
- Use `embed` for HTML templates
- Use `flag` for CLI args
- Tests: `go test ./...`
- Format: `gofmt`
