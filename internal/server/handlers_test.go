package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleIndex(t *testing.T) {
	dir := setupTestLogDir(t)
	srv := New(dir)

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !containsString(body, "Projects") {
		t.Error("response should contain 'Projects'")
	}
	if !containsString(body, "/Users/foo/workspace/proj") {
		t.Error("response should contain decoded project path")
	}
}

func TestHandleProject(t *testing.T) {
	dir := setupTestLogDir(t)
	srv := New(dir)

	req := httptest.NewRequest("GET", "/projects/-Users-foo-workspace-proj", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !containsString(body, "Hello from test") {
		t.Error("response should contain first message summary")
	}
}

func TestHandleSession(t *testing.T) {
	dir := setupTestLogDir(t)
	srv := New(dir)

	req := httptest.NewRequest("GET", "/sessions/-Users-foo-workspace-proj/sess-1", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !containsString(body, "Hello from test") {
		t.Error("response should contain user message")
	}
	if !containsString(body, "claude-opus-4-6") {
		t.Error("response should contain model name")
	}
}

func TestHandleNotFound(t *testing.T) {
	dir := setupTestLogDir(t)
	srv := New(dir)

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func setupTestLogDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	projDir := filepath.Join(dir, "-Users-foo-workspace-proj")
	os.MkdirAll(projDir, 0755)

	content := `{"type":"user","uuid":"u1","timestamp":"2026-02-25T06:41:55.945Z","sessionId":"sess-1","message":{"role":"user","content":"Hello from test"}}
{"type":"assistant","uuid":"a1","timestamp":"2026-02-25T06:42:02.218Z","sessionId":"sess-1","message":{"model":"claude-opus-4-6","role":"assistant","content":[{"type":"text","text":"Hi there!"}],"usage":{"input_tokens":100,"output_tokens":50,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}}}
`
	os.WriteFile(filepath.Join(projDir, "sess-1.jsonl"), []byte(content), 0644)
	return dir
}

func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && // guard
		len(s) >= len(substr) &&
		(s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
