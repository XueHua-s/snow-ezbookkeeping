package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAIAssistantPreferredReplyLanguage(t *testing.T) {
	assert.Equal(t, "Simplified Chinese", getAIAssistantPreferredReplyLanguage("zh-CN"))
	assert.Equal(t, "Traditional Chinese", getAIAssistantPreferredReplyLanguage("zh-Hant"))
	assert.Equal(t, "Japanese", getAIAssistantPreferredReplyLanguage("ja-JP"))
	assert.Equal(t, "English", getAIAssistantPreferredReplyLanguage("en-US"))
	assert.Equal(t, "English", getAIAssistantPreferredReplyLanguage(""))
}
