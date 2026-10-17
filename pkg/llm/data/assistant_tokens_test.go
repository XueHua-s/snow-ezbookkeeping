package data

import (
	"strings"
	"testing"
)

func TestAssistantTokenBudgetIsIndependentOfUTF8Bytes(t *testing.T) {
	raw := []byte(`{"content":"` + strings.Repeat("日常消费预算", 1000) + `"}`)
	tokens := EstimateAssistantTokens(string(raw))
	if tokens >= len(raw) || tokens < 1000 {
		t.Fatal("unexpected multilingual estimate", tokens, len(raw))
	}
	if !AssistantRequestFits(raw, tokens, len(raw)) {
		t.Fatal("exact budgets rejected")
	}
	if AssistantRequestFits(raw, tokens-1, len(raw)) {
		t.Fatal("token budget bypassed")
	}
	if AssistantRequestFits(raw, tokens, len(raw)-1) {
		t.Fatal("byte budget bypassed")
	}
	if AssistantRequestFits([]byte(strings.Repeat("a", AssistantMaxRequestBytes+1)), 0, AssistantMaxRequestBytes*2) {
		t.Fatal("hard byte cap bypassed")
	}
}

func TestAssistantTokenizerConcurrentAndBoundedLongStrings(t *testing.T) {
	for _, s := range []string{strings.Repeat("a", 128*1024), strings.Repeat("中文😀", 20000)} {
		t.Run(s[:1], func(t *testing.T) {
			t.Parallel()
			expected := EstimateAssistantTokens(s)
			if expected <= 0 || expected > len(s) {
				t.Fatal(expected)
			}
			for i := 0; i < 4; i++ {
				if EstimateAssistantTokens(s) != expected {
					t.Fatal("unstable token count")
				}
			}
		})
	}
}
