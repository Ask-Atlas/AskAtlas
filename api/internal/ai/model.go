package ai

import "github.com/openai/openai-go/shared"

// Model is an OpenAI chat model identifier. Aliased to the SDK's
// ChatModel type so callers don't need to import openai-go directly.
type Model = shared.ChatModel

// Model defaults. The GPT-4.1 family is the current production tier
// in openai-go v1.12: 4.1 is the flagship, 4.1-mini is the cost-
// balanced workhorse, 4.1-nano is the cheapest path for verification
// and short classification.
//
// Pinned to the un-dated aliases ("gpt-4.1") so OpenAI can roll
// forward without us pushing code; switch to a dated SHA if a
// regression appears.
const (
	ModelGPT41     Model = shared.ChatModelGPT4_1
	ModelGPT41Mini Model = shared.ChatModelGPT4_1Mini
	ModelGPT41Nano Model = shared.ChatModelGPT4_1Nano

	// ModelDefault is the quality tier used by edits, grounded gen,
	// and Q&A. Override per-call to ModelCheap for short tasks.
	ModelDefault = ModelGPT41
	// ModelCheap is the entry-level alias used by smoke + classifier
	// tasks; cost is roughly 5% of ModelDefault.
	ModelCheap = ModelGPT41Nano
)

// MaxTokensDefault caps response length when the caller leaves
// StreamRequest.MaxTokens at zero. 1024 is enough for the smoke
// endpoint and short edits; long-form generation should override.
const MaxTokensDefault int64 = 1024
