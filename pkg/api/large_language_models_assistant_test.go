package api

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func TestGetAIAssistantPreferredReplyLanguage(t *testing.T) {
	assert.Equal(t, "Simplified Chinese", getAIAssistantPreferredReplyLanguage("zh-CN"))
	assert.Equal(t, "Traditional Chinese", getAIAssistantPreferredReplyLanguage("zh-Hant"))
	assert.Equal(t, "Japanese", getAIAssistantPreferredReplyLanguage("ja-JP"))
	assert.Equal(t, "English", getAIAssistantPreferredReplyLanguage("en-US"))
	assert.Equal(t, "English", getAIAssistantPreferredReplyLanguage(""))
}

func TestBuildAIAssistantUserPrompt_SummaryUsesChinesePromptForChineseLocale(t *testing.T) {
	api := &LargeLanguageModelsApi{}
	request := &models.AIAssistantChatRequest{
		Message: "重点看餐饮支出",
	}

	prompt := api.buildAIAssistantUserPrompt(request, models.AIAssistantModeSummary, "zh-CN")
	assert.Contains(t, prompt, "请基于我的账单数据，用中文给我一份个人财务总结")
	assert.Contains(t, prompt, "额外关注点：重点看餐饮支出")
}

func TestBuildAIAssistantUserPrompt_SummaryUsesEnglishPromptForEnglishLocale(t *testing.T) {
	api := &LargeLanguageModelsApi{}
	request := &models.AIAssistantChatRequest{
		Message: "Focus on food spending",
	}

	prompt := api.buildAIAssistantUserPrompt(request, models.AIAssistantModeSummary, "en-US")
	assert.Contains(t, prompt, "Please provide a personal finance summary")
	assert.Contains(t, prompt, "Additional focus: Focus on food spending")
}

func TestBuildAIAssistantEmbeddingQueryText_SummaryUsesChineseQueryForChineseLocale(t *testing.T) {
	api := &LargeLanguageModelsApi{}
	request := &models.AIAssistantChatRequest{
		Message: "重点看餐饮支出",
	}

	queryText := api.buildAIAssistantEmbeddingQueryText(request, models.AIAssistantModeSummary, "zh-CN")
	assert.Equal(t, "个人财务总结与记账建议，重点关注：重点看餐饮支出", queryText)
}

func TestBuildAIAssistantEmbeddingQueryText_SummaryUsesEnglishQueryForEnglishLocale(t *testing.T) {
	api := &LargeLanguageModelsApi{}
	request := &models.AIAssistantChatRequest{}

	queryText := api.buildAIAssistantEmbeddingQueryText(request, models.AIAssistantModeSummary, "en-US")
	assert.Equal(t, "summarize recent personal finance trends, spending, risks, and bookkeeping suggestions", queryText)
}
