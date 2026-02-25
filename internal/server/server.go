package server

import (
	"encoding/json"
	"html/template"
	"log/slog"
	"net/http"

	"github.com/nhosoya/claude-code-share/internal/logparser"
	"github.com/nhosoya/claude-code-share/internal/templates"
)

// Server holds the HTTP server configuration.
type Server struct {
	LogDir string
	pages  map[string]*template.Template
}

// New creates a new Server with parsed templates.
func New(logDir string) *Server {
	funcMap := template.FuncMap{
		"formatToolInput": formatToolInput,
		"hasText":         hasText,
	}

	// Parse each page template together with the layout so that
	// "title" and "content" blocks don't collide across pages.
	pageNames := []string{"index.html", "project.html", "session.html"}
	pages := make(map[string]*template.Template, len(pageNames))
	for _, name := range pageNames {
		pages[name] = template.Must(
			template.New("").Funcs(funcMap).ParseFS(templates.FS, "layout.html", name),
		)
	}

	return &Server{
		LogDir: logDir,
		pages:  pages,
	}
}

// Handler returns an http.Handler with all routes configured.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/projects/", s.handleProject)
	mux.HandleFunc("/sessions/", s.handleSession)
	return mux
}

func (s *Server) render(w http.ResponseWriter, page string, data interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := s.pages[page]
	if tmpl == nil {
		slog.Error("template not found", "page", page)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		slog.Error("template render error", "page", page, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// hasText returns true if an assistant message contains at least one text block.
func hasText(blocks []logparser.ContentBlock) bool {
	for _, b := range blocks {
		if b.Type == "text" && b.Text != "" {
			return true
		}
	}
	return false
}

func formatToolInput(input map[string]interface{}) string {
	b, err := json.MarshalIndent(input, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}
