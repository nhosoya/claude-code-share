package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/nhosoya/claude-code-share/internal/logparser"
)

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	projects, err := logparser.ListProjects(s.LogDir)
	if err != nil {
		slog.Error("failed to list projects", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.render(w, "index.html", struct {
		Projects []logparser.Project
	}{Projects: projects})
}

func (s *Server) handleProject(w http.ResponseWriter, r *http.Request) {
	// URL: /projects/{slug}
	slug := strings.TrimPrefix(r.URL.Path, "/projects/")
	slug = strings.TrimSuffix(slug, "/")
	if slug == "" {
		http.NotFound(w, r)
		return
	}

	sessions, err := logparser.ListSessions(s.LogDir, slug)
	if err != nil {
		slog.Error("failed to list sessions", "error", err, "slug", slug)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	s.render(w, "project.html", struct {
		Slug     string
		Path     string
		Sessions []logparser.Session
	}{
		Slug:     slug,
		Path:     logparser.DecodeSlug(slug),
		Sessions: sessions,
	})
}

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	// URL: /sessions/{slug}/{sessionId}
	parts := strings.TrimPrefix(r.URL.Path, "/sessions/")
	parts = strings.TrimSuffix(parts, "/")
	idx := strings.Index(parts, "/")
	if idx < 0 {
		http.NotFound(w, r)
		return
	}
	slug := parts[:idx]
	sessionID := parts[idx+1:]
	if slug == "" || sessionID == "" {
		http.NotFound(w, r)
		return
	}

	conv, err := logparser.LoadSession(s.LogDir, slug, sessionID)
	if err != nil {
		slog.Error("failed to load session", "error", err, "slug", slug, "session", sessionID)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	s.render(w, "session.html", struct {
		Slug         string
		Path         string
		SessionID    string
		Conversation *logparser.Conversation
	}{
		Slug:         slug,
		Path:         logparser.DecodeSlug(slug),
		SessionID:    sessionID,
		Conversation: conv,
	})
}
