# Initial Build Prompt

Copy the prompt below and paste it into a new Claude Code session opened in this repository's directory.

---

## Prompt

Build the `claude-code-share` CLI tool following the spec in `CLAUDE.md`.

### Goal

A single Go binary that reads Claude Code JSONL conversation logs from `~/.claude/projects/` and serves them as a browsable web UI. The purpose is to let team members on the same network view each other's Claude Code sessions to learn how others use it.

### Requirements

#### CLI

- `claude-code-share` with no args starts the server
- `--port` (default: `3333`)
- `--host` (default: `0.0.0.0` — accessible from LAN)
- `--log-dir` (default: `~/.claude/projects`)
- On startup, print the local URL and any LAN-accessible URLs (detect network interfaces)

#### Pages

1. **`/` — Project List**
   - List all project directories found under `--log-dir`
   - Show the decoded workspace path (reverse the slug: `-Users-foo-workspace-my-project` → `/Users/foo/workspace/my-project`)
   - Show session count and last activity timestamp per project
   - Sort by last activity (newest first)

2. **`/projects/{slug}` — Session List**
   - List all `.jsonl` sessions for the selected project
   - Show: first user message (as summary), timestamp, message count, model used
   - Sort by timestamp (newest first)

3. **`/sessions/{slug}/{sessionId}` — Conversation View**
   - Render the full conversation in a chat-like layout
   - **User messages**: show the text content. When `message.content` is an array (tool results), show them in a collapsed/summary form
   - **Assistant messages**: iterate over the `content` array
     - `type: "text"` → render as Markdown (use a simple Markdown-to-HTML converter or just preserve whitespace and line breaks)
     - `type: "tool_use"` → show tool name and a collapsible detail section with the input parameters
   - **Skip** `progress` and `file-history-snapshot` entries — these add noise
   - Show token usage summary (total input/output tokens) at the top of the page
   - Show model name and timestamps

#### UI/Design

- Clean, minimal design with a monospace font
- Light background, good readability
- No JavaScript frameworks — server-rendered HTML with minimal inline CSS
- Responsive enough to be readable on mobile
- Collapsible sections for tool call details (use `<details>/<summary>` HTML elements)
- Navigation breadcrumbs: Home > Project > Session

#### Code Quality

- Write tests for the JSONL parser (`internal/logparser/`)
- Use `go:embed` for templates
- Handle malformed JSONL lines gracefully (skip and log warning)
- Keep external dependencies to zero (stdlib only)

### Steps

1. `go mod init github.com/nhosoya/claude-code-share`
2. Implement `internal/logparser/models.go` — data structures
3. Implement `internal/logparser/parser.go` — JSONL reading and parsing
4. Write tests for the parser
5. Implement `internal/templates/` — HTML templates with `embed`
6. Implement `internal/server/` — HTTP handlers
7. Implement `main.go` — CLI flags and server startup
8. Run `go build` and verify it works
9. Write a short `README.md` with usage instructions

Use TDD where practical: write parser tests before or alongside the implementation.
