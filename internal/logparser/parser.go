package logparser

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DecodeSlug converts a project directory slug back to the original workspace path.
// Slug format: leading `-` becomes `/`, then each `-` becomes `/`,
// but `.` in the original path was kept as-is.
//
// The slug is formed by replacing `/` with `-` in the absolute path,
// which means the leading `/` becomes a leading `-`.
func DecodeSlug(slug string) string {
	if len(slug) == 0 {
		return ""
	}
	// The slug starts with `-` which represents the leading `/`.
	// Each `-` in the slug represents a `/` in the original path.
	// Dots are preserved as-is.
	return strings.ReplaceAll(slug, "-", "/")
}

// ParseEntry parses a single JSONL line into a LogEntry.
func ParseEntry(line []byte) (LogEntry, error) {
	var entry LogEntry
	if err := json.Unmarshal(line, &entry); err != nil {
		return LogEntry{}, fmt.Errorf("parse entry: %w", err)
	}

	// Parse the content field which can be string or array
	switch raw := entry.Message.RawContent.(type) {
	case string:
		entry.Message.Content = MessageContent{Text: raw}
	case []interface{}:
		data, _ := json.Marshal(raw)
		var blocks []ContentBlock
		if err := json.Unmarshal(data, &blocks); err != nil {
			slog.Warn("failed to parse content blocks", "error", err)
		}
		entry.Message.Content = MessageContent{Blocks: blocks}
	}

	return entry, nil
}

// ParseSessionFile reads a JSONL file and returns a Conversation.
// Skips malformed lines and progress/file-history-snapshot entries.
func ParseSessionFile(path string) (*Conversation, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open session file: %w", err)
	}
	defer f.Close()

	conv := &Conversation{
		SessionID: strings.TrimSuffix(filepath.Base(path), ".jsonl"),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 10*1024*1024) // 10MB max line size
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		entry, err := ParseEntry(line)
		if err != nil {
			slog.Warn("skipping malformed JSONL line", "error", err, "file", path)
			continue
		}

		// Skip noise entries
		if entry.Type == "progress" || entry.Type == "file-history-snapshot" {
			continue
		}

		// Accumulate token usage
		if entry.Message.Usage != nil {
			conv.TotalInput += entry.Message.Usage.InputTokens
			conv.TotalOutput += entry.Message.Usage.OutputTokens
		}

		// Capture model name from first assistant message
		if conv.Model == "" && entry.Message.Model != "" {
			conv.Model = entry.Message.Model
		}

		conv.Entries = append(conv.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan session file: %w", err)
	}

	return conv, nil
}

// ListProjects scans the log directory for project subdirectories.
func ListProjects(logDir string) ([]Project, error) {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil, fmt.Errorf("read log dir: %w", err)
	}

	var projects []Project
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		slug := e.Name()
		projDir := filepath.Join(logDir, slug)

		sessionFiles, _ := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
		if len(sessionFiles) == 0 {
			continue
		}

		var lastActivity time.Time
		for _, sf := range sessionFiles {
			info, err := os.Stat(sf)
			if err != nil {
				continue
			}
			if info.ModTime().After(lastActivity) {
				lastActivity = info.ModTime()
			}
		}

		projects = append(projects, Project{
			Slug:         slug,
			Path:         DecodeSlug(slug),
			SessionCount: len(sessionFiles),
			LastActivity: lastActivity,
		})
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastActivity.After(projects[j].LastActivity)
	})

	return projects, nil
}

// ListSessions returns all sessions for a given project slug.
func ListSessions(logDir, slug string) ([]Session, error) {
	projDir := filepath.Join(logDir, slug)
	sessionFiles, err := filepath.Glob(filepath.Join(projDir, "*.jsonl"))
	if err != nil {
		return nil, fmt.Errorf("glob sessions: %w", err)
	}

	var sessions []Session
	for _, sf := range sessionFiles {
		conv, err := ParseSessionFile(sf)
		if err != nil {
			slog.Warn("skipping session file", "error", err, "file", sf)
			continue
		}

		sess := Session{
			ID:           conv.SessionID,
			MessageCount: len(conv.Entries),
			Model:        conv.Model,
		}

		// Find first user message and timestamp
		for _, e := range conv.Entries {
			if sess.Timestamp.IsZero() {
				sess.Timestamp = e.Timestamp
			}
			if e.Type == "user" && sess.FirstMessage == "" {
				text := e.Message.Content.Text
				if text == "" && len(e.Message.Content.Blocks) > 0 {
					text = "(tool results)"
				}
				sess.FirstMessage = truncate(text, 120)
			}
			if sess.FirstMessage != "" && !sess.Timestamp.IsZero() {
				break
			}
		}

		sessions = append(sessions, sess)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].Timestamp.After(sessions[j].Timestamp)
	})

	return sessions, nil
}

// LoadSession loads a specific session file by project slug and session ID.
func LoadSession(logDir, slug, sessionID string) (*Conversation, error) {
	path := filepath.Join(logDir, slug, sessionID+".jsonl")
	return ParseSessionFile(path)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
