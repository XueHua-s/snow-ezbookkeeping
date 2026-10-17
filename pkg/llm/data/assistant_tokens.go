package data

import (
	"sync"
	"unicode/utf8"

	"github.com/tiktoken-go/tokenizer/codec"
)

const AssistantMaxRequestBytes = 128 * 1024

var assistantTokenizer = sync.OnceValue(codec.NewCl100kBase)

// EstimateAssistantTokens uses an embedded BPE vocabulary with 20% headroom.
// It is a portable estimate, not a guarantee for every configured model tokenizer.
// Counting serialized input includes schemas and opaque provider continuation state;
// byte limits remain independent and are checked before tokenization.
func EstimateAssistantTokens(text string) int {
	// Bound BPE's work on long unbroken strings (including encrypted state).
	// Chunk boundaries can add a few tokens, which is safe for this estimate.
	count := 0
	for len(text) > 0 {
		end := min(len(text), 2048)
		for end < len(text) && end > 0 && !utf8.RuneStart(text[end]) {
			end--
		}
		if end == 0 {
			end = min(len(text), 2048)
		}
		n, err := assistantTokenizer().Count(text[:end])
		if err != nil {
			n = end
		}
		count += n
		text = text[end:]
	}
	return (count*6 + 4) / 5
}

// AssistantRequestFits applies the independent transport and estimated token budgets.
// A zero token budget is unlimited for legacy callers; the transport cap always applies.
func AssistantRequestFits(raw []byte, tokens, bytes int) bool {
	if bytes <= 0 || bytes > AssistantMaxRequestBytes {
		bytes = AssistantMaxRequestBytes
	}
	return len(raw) <= bytes && (tokens <= 0 || EstimateAssistantTokens(string(raw)) <= tokens)
}
