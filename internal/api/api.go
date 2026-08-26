// Package api defines the JSON types shared between the ws CLI and
// the ws-server HTTP API.
package api

import "time"

const MaxResponseBytes = 64 << 20

// Origin records where an entry was created: the OS user, host, and
// working directory of the ws client, plus the Claude Code session id
// when one is set. Host and dir are always captured; user and the
// Claude session may be empty when they cannot be determined.
type Origin struct {
	User          string `json:"user,omitempty"`
	Host          string `json:"host,omitempty"`
	Dir           string `json:"dir,omitempty"`
	ClaudeSession string `json:"claude_session,omitempty"`
}

type Entry struct {
	ID       int64             `json:"id"`
	Created  time.Time         `json:"created"`
	Modified time.Time         `json:"modified"`
	Type     string            `json:"type"`
	Subject  string            `json:"subject"`
	Body     string            `json:"body,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Origin   Origin            `json:"origin"`
}

type AddEntryRequest struct {
	Type     string            `json:"type"`
	Subject  string            `json:"subject"`
	Body     string            `json:"body,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
	Origin   Origin            `json:"origin"`
}

// EditEntryRequest uses pointers so a field can be updated to the
// empty string (e.g. clearing the body) while absent fields are left
// untouched. A metadata key given an empty value is removed.
type EditEntryRequest struct {
	Type     *string           `json:"type,omitempty"`
	Subject  *string           `json:"subject,omitempty"`
	Body     *string           `json:"body,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type SearchResult struct {
	Entries []Entry `json:"entries"`
	Total   int     `json:"total"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type StatusResponse struct {
	Version        string `json:"version"`
	Database       string `json:"database"`
	Data           string `json:"data"`
	Timeout        string `json:"timeout"`
	Authentication bool   `json:"authentication"`
}
