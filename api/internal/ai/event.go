// Package ai provides an OpenAI-backed LLM client and an SSE helper
// so feature handlers can stream Chat-Completions responses to
// clients without re-implementing prompt-cache plumbing, cost
// logging, or event framing.
//
// The package surfaces three layers:
//
//   - Request / Event / Usage types -- the wire-shape contract any
//     caller agrees to.
//   - Client (openai-go wrapper) -- produces a <-chan Event that
//     closes on stream end / cancellation.
//   - SSE helper -- writes a <-chan Event to a tmaxmax/go-sse Session
//     using the typed event names below.
//
// Feature endpoints (ASK-215 edit, ASK-225 Q&A, ASK-226 quiz, etc.)
// build on top of these primitives so they share the same auth, cost
// log, and cancellation behavior. ASK-213 ships the smoke endpoint
// (POST /api/ai/ping) that exercises the whole chain.
//
// On prompt caching: OpenAI applies caching automatically for
// prompts >= 1024 tokens (no marker required). The Block.CacheControl
// field is therefore a hint-only signal -- callers should still place
// long stable content (system prompt, retrieved chunks) FIRST in the
// message so the auto-cache windows match across requests, but no
// API marker is sent.
package ai

import "github.com/google/uuid"

// EventKind names the SSE `event:` field values produced by Stream.
// Any change here is a wire-protocol change -- bump it deliberately.
type EventKind string

const (
	// EventDelta carries an incremental text chunk from the model.
	// Many of these arrive per request.
	EventDelta EventKind = "delta"
	// EventUsage carries the final token + cache counts. Emitted
	// exactly once, just before EventDone.
	EventUsage EventKind = "usage"
	// EventError is emitted at most once when the upstream stream
	// fails mid-flight. The handler closes the channel after it.
	EventError EventKind = "error"
	// EventDone is the terminal sentinel. Always the last event on
	// a successful stream; absent on streams that errored.
	EventDone EventKind = "done"
)

// Event is the union of all values flowing through Stream's channel.
// Exactly one payload field is set per event; the others are zero.
type Event struct {
	Kind  EventKind
	Delta string
	Usage *Usage
	Err   error
}

// Usage captures the token + cache counts a single request consumed.
// Emitted once at end of stream and also passed to the cost log so
// ASK-214's ledger has the same numbers the client saw.
type Usage struct {
	InputTokens      int64 `json:"input_tokens"`
	OutputTokens     int64 `json:"output_tokens"`
	CacheReadTokens  int64 `json:"cache_read_tokens"`
	CacheWriteTokens int64 `json:"cache_write_tokens"`
}

// Feature names the calling product surface for cost-log + future
// rate-limiting attribution. Free-form string typed for clarity at
// the call site; the cost log writes it as-is.
type Feature string

const (
	FeaturePing         Feature = "ping"
	FeatureEdit         Feature = "edit"
	FeatureGroundedEdit Feature = "grounded_edit"
	FeatureQA           Feature = "qa"
	FeatureQuiz         Feature = "quiz"
	FeatureRefSuggest   Feature = "ref_suggest"
)

// StreamRequest is the input to Client.Stream. Mirrors the subset of
// openai-go's ChatCompletionNewParams the rest of the codebase needs;
// callers do not import the SDK directly.
//
// System blocks land in a single "system" Chat-Completions message;
// each Block becomes a content part. The CacheControl field is
// hint-only on OpenAI (provider auto-caches) -- see Block doc.
type StreamRequest struct {
	// UserID attributes cost + usage to the caller. Required.
	UserID uuid.UUID
	// Feature names the calling surface for cost-log + (future) rate
	// limiting. Required.
	Feature Feature
	// Model overrides the default. Zero value = ModelDefault (Sonnet).
	Model Model
	// MaxTokens caps response length. Zero value = MaxTokensDefault.
	MaxTokens int64
	// System blocks. Each block is sent as a separate `text` block
	// so individual blocks can carry cache_control.
	System []Block
	// Messages is the conversation history ending in the user turn.
	// Empty Messages is rejected.
	Messages []Message
}

// Block is a piece of text in either a system prompt or a user turn.
// CacheControl is a hint-only flag carried over from the original
// Anthropic plumbing -- OpenAI auto-caches prompts >= 1024 tokens
// without an explicit marker, so this field is a no-op on the wire.
// Kept on the struct so callers can still reason about which blocks
// are "long + stable" (place those first to maximise auto-cache
// hits); a future provider switch can re-enable explicit markers.
type Block struct {
	Text         string
	CacheControl bool
}

// Role is the speaker of a Message. Only "user" and "assistant" are
// valid; system text goes in StreamRequest.System.
type Role string

const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
)

// Message is one turn in the conversation. Multi-block messages let
// callers stitch retrieved chunks + a question into a single user
// turn while keeping the chunks individually cacheable.
type Message struct {
	Role   Role
	Blocks []Block
}

// CompleteResponse is the result of a non-streaming Complete call.
// Used by features that don't need token-by-token UI updates --
// e.g. ASK-228 (rank candidate entity refs) or any classifier.
type CompleteResponse struct {
	Text  string
	Usage Usage
}
