package openai

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/core"
)

func TestOpenAIOfficialChatCompletionsAPIProvider_BuildChatCompletionsHttpRequest(t *testing.T) {
	apiProvider := &OpenAIOfficialChatCompletionsAPIProvider{
		OpenAIAPIKey:             "test-key",
		OpenAIChatCompletionsURL: "https://api.example.com/v1/chat/completions",
	}

	req, err := apiProvider.BuildChatCompletionsHttpRequest(core.NewNullContext(), 0)
	assert.Nil(t, err)
	assert.Equal(t, "POST", req.Method)
	assert.Equal(t, "https://api.example.com/v1/chat/completions", req.URL.String())
	assert.Equal(t, "Bearer test-key", req.Header.Get("Authorization"))
}
