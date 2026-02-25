package logparser

import "time"

// LogEntry represents a single line from a JSONL log file.
type LogEntry struct {
	Type       string    `json:"type"`
	UUID       string    `json:"uuid"`
	ParentUUID *string   `json:"parentUuid"`
	Timestamp  time.Time `json:"timestamp"`
	SessionID  string    `json:"sessionId"`
	Version    string    `json:"version,omitempty"`
	CWD        string    `json:"cwd,omitempty"`
	Message    Message   `json:"message"`
}

// Message represents the message field in a log entry.
type Message struct {
	Role    string         `json:"role"`
	Model   string         `json:"model,omitempty"`
	Content MessageContent `json:"-"`
	Usage   *Usage         `json:"usage,omitempty"`

	// RawContent holds the raw JSON for deferred parsing of content.
	RawContent interface{} `json:"content"`
}

// MessageContent can be either a string or an array of content blocks.
type MessageContent struct {
	Text   string         // When content is a plain string
	Blocks []ContentBlock // When content is an array
}

// ContentBlock represents a single element in an assistant's content array.
type ContentBlock struct {
	Type  string                 `json:"type"`
	Text  string                 `json:"text,omitempty"`
	ID    string                 `json:"id,omitempty"`
	Name  string                 `json:"name,omitempty"`
	Input map[string]interface{} `json:"input,omitempty"`

	// For tool_result blocks in user messages
	ToolUseID string      `json:"tool_use_id,omitempty"`
	Content   interface{} `json:"content,omitempty"`
}

// Usage tracks token consumption.
type Usage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
}

// Project represents a project directory containing sessions.
type Project struct {
	Slug         string
	Path         string // Decoded workspace path
	SessionCount int
	LastActivity time.Time
}

// Session represents a single conversation session.
type Session struct {
	ID           string
	FirstMessage string
	Timestamp    time.Time
	MessageCount int
	Model        string
}

// Conversation holds all entries for a single session view.
type Conversation struct {
	SessionID   string
	Entries     []LogEntry
	TotalInput  int
	TotalOutput int
	Model       string
}
