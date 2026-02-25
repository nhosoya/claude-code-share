package logparser

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDecodeSlug(t *testing.T) {
	// Note: slug encoding is lossy — both `/` and `.` in the original path
	// become `-`. We do best-effort decode by replacing `-` with `/`.
	tests := []struct {
		slug string
		want string
	}{
		{"-Users-foo-workspace-proj", "/Users/foo/workspace/proj"},
		{"-home-user-code", "/home/user/code"},
		{"", ""},
	}
	for _, tt := range tests {
		got := DecodeSlug(tt.slug)
		if got != tt.want {
			t.Errorf("DecodeSlug(%q) = %q, want %q", tt.slug, got, tt.want)
		}
	}
}

func TestParseEntry_UserText(t *testing.T) {
	line := `{"type":"user","uuid":"abc","timestamp":"2026-02-25T06:41:55.945Z","sessionId":"sess-1","message":{"role":"user","content":"Hello world"}}`
	entry, err := ParseEntry([]byte(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Type != "user" {
		t.Errorf("Type = %q, want %q", entry.Type, "user")
	}
	if entry.Message.Content.Text != "Hello world" {
		t.Errorf("Content.Text = %q, want %q", entry.Message.Content.Text, "Hello world")
	}
	if len(entry.Message.Content.Blocks) != 0 {
		t.Errorf("Content.Blocks should be empty for text content")
	}
}

func TestParseEntry_AssistantWithToolUse(t *testing.T) {
	line := `{"type":"assistant","uuid":"def","timestamp":"2026-02-25T06:42:02.218Z","sessionId":"sess-1","message":{"model":"claude-opus-4-6","role":"assistant","content":[{"type":"text","text":"Here is the result"},{"type":"tool_use","id":"toolu_01","name":"Bash","input":{"command":"ls"}}],"usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":200,"cache_read_input_tokens":0}}}`
	entry, err := ParseEntry([]byte(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entry.Type != "assistant" {
		t.Errorf("Type = %q, want %q", entry.Type, "assistant")
	}
	if entry.Message.Model != "claude-opus-4-6" {
		t.Errorf("Model = %q, want %q", entry.Message.Model, "claude-opus-4-6")
	}
	if len(entry.Message.Content.Blocks) != 2 {
		t.Fatalf("Content.Blocks length = %d, want 2", len(entry.Message.Content.Blocks))
	}
	if entry.Message.Content.Blocks[0].Type != "text" {
		t.Errorf("Block[0].Type = %q, want %q", entry.Message.Content.Blocks[0].Type, "text")
	}
	if entry.Message.Content.Blocks[1].Name != "Bash" {
		t.Errorf("Block[1].Name = %q, want %q", entry.Message.Content.Blocks[1].Name, "Bash")
	}
	if entry.Message.Usage == nil {
		t.Fatal("Usage should not be nil")
	}
	if entry.Message.Usage.InputTokens != 100 {
		t.Errorf("InputTokens = %d, want 100", entry.Message.Usage.InputTokens)
	}
}

func TestParseEntry_UserWithToolResults(t *testing.T) {
	line := `{"type":"user","uuid":"ghi","timestamp":"2026-02-25T06:43:00.000Z","sessionId":"sess-1","message":{"role":"user","content":[{"type":"tool_result","tool_use_id":"toolu_01","content":"file list output"}]}}`
	entry, err := ParseEntry([]byte(line))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entry.Message.Content.Blocks) != 1 {
		t.Fatalf("Content.Blocks length = %d, want 1", len(entry.Message.Content.Blocks))
	}
	if entry.Message.Content.Blocks[0].Type != "tool_result" {
		t.Errorf("Block[0].Type = %q, want %q", entry.Message.Content.Blocks[0].Type, "tool_result")
	}
}

func TestParseEntry_SkippedTypes(t *testing.T) {
	for _, typ := range []string{"progress", "file-history-snapshot"} {
		line := `{"type":"` + typ + `","uuid":"x","timestamp":"2026-02-25T06:41:55.945Z","sessionId":"sess-1","message":{}}`
		entry, err := ParseEntry([]byte(line))
		if err != nil {
			t.Fatalf("unexpected error for type %q: %v", typ, err)
		}
		if entry.Type != typ {
			t.Errorf("Type = %q, want %q", entry.Type, typ)
		}
	}
}

func TestParseEntry_MalformedJSON(t *testing.T) {
	_, err := ParseEntry([]byte(`{bad json`))
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseSessionFile(t *testing.T) {
	dir := t.TempDir()
	content := `{"type":"user","uuid":"a","timestamp":"2026-02-25T06:41:55.945Z","sessionId":"sess-1","message":{"role":"user","content":"Hi"}}
{"type":"assistant","uuid":"b","timestamp":"2026-02-25T06:42:02.218Z","sessionId":"sess-1","message":{"model":"claude-opus-4-6","role":"assistant","content":[{"type":"text","text":"Hello!"}],"usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
{bad line that should be skipped}
{"type":"progress","uuid":"c","timestamp":"2026-02-25T06:42:03.000Z","sessionId":"sess-1","message":{}}
`
	path := filepath.Join(dir, "sess-1.jsonl")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	conv, err := ParseSessionFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// progress and malformed lines should be filtered/skipped
	if len(conv.Entries) != 2 {
		t.Errorf("Entries length = %d, want 2 (user + assistant, skipping progress and malformed)", len(conv.Entries))
	}
	if conv.TotalInput != 10 {
		t.Errorf("TotalInput = %d, want 10", conv.TotalInput)
	}
	if conv.TotalOutput != 5 {
		t.Errorf("TotalOutput = %d, want 5", conv.TotalOutput)
	}
	if conv.Model != "claude-opus-4-6" {
		t.Errorf("Model = %q, want %q", conv.Model, "claude-opus-4-6")
	}
}

func TestListProjects(t *testing.T) {
	dir := t.TempDir()

	// Create two project directories
	proj1 := filepath.Join(dir, "-Users-foo-workspace-proj1")
	proj2 := filepath.Join(dir, "-Users-foo-workspace-proj2")
	os.MkdirAll(proj1, 0755)
	os.MkdirAll(proj2, 0755)

	// proj1: 2 sessions, older
	writeTestSession(t, proj1, "sess-a.jsonl", "2026-02-24T10:00:00.000Z", "First message")
	writeTestSession(t, proj1, "sess-b.jsonl", "2026-02-24T12:00:00.000Z", "Second message")

	// proj2: 1 session, newer
	writeTestSession(t, proj2, "sess-c.jsonl", "2026-02-25T06:00:00.000Z", "Newer message")

	projects, err := ListProjects(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("projects length = %d, want 2", len(projects))
	}
	// Should be sorted newest first
	if projects[0].Slug != "-Users-foo-workspace-proj2" {
		t.Errorf("first project slug = %q, want %q", projects[0].Slug, "-Users-foo-workspace-proj2")
	}
	if projects[0].SessionCount != 1 {
		t.Errorf("first project session count = %d, want 1", projects[0].SessionCount)
	}
	if projects[1].SessionCount != 2 {
		t.Errorf("second project session count = %d, want 2", projects[1].SessionCount)
	}
}

func TestListSessions(t *testing.T) {
	dir := t.TempDir()
	projDir := filepath.Join(dir, "-Users-foo-workspace-proj1")
	os.MkdirAll(projDir, 0755)

	writeTestSession(t, projDir, "sess-a.jsonl", "2026-02-24T10:00:00.000Z", "Hello from session A")
	writeTestSession(t, projDir, "sess-b.jsonl", "2026-02-25T06:00:00.000Z", "Hello from session B")

	sessions, err := ListSessions(dir, "-Users-foo-workspace-proj1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("sessions length = %d, want 2", len(sessions))
	}
	// Newest first
	if sessions[0].ID != "sess-b" {
		t.Errorf("first session ID = %q, want %q", sessions[0].ID, "sess-b")
	}
	if sessions[0].FirstMessage != "Hello from session B" {
		t.Errorf("first session FirstMessage = %q, want %q", sessions[0].FirstMessage, "Hello from session B")
	}
}

func writeTestSession(t *testing.T, dir, filename, timestamp, firstMsg string) {
	t.Helper()
	content := `{"type":"user","uuid":"u1","timestamp":"` + timestamp + `","sessionId":"test","message":{"role":"user","content":"` + firstMsg + `"}}
{"type":"assistant","uuid":"a1","timestamp":"` + timestamp + `","sessionId":"test","message":{"model":"claude-opus-4-6","role":"assistant","content":[{"type":"text","text":"OK"}],"usage":{"input_tokens":10,"output_tokens":5,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
`
	path := filepath.Join(dir, filename)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	// Set modification time based on timestamp for consistent ordering
	ts, _ := time.Parse(time.RFC3339Nano, timestamp)
	os.Chtimes(path, ts, ts)
}
