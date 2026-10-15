package api

import (
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func normalizeAIAssistantMode(mode string) (string, error) {
	if mode == "" {
		return models.AIAssistantModeChat, nil
	}

	if mode == models.AIAssistantModeChat || mode == models.AIAssistantModeSummary {
		return mode, nil
	}

	return "", errs.ErrAIAssistantInvalidMode
}

func normalizeAIAssistantClientLocale(clientLocale string) string {
	normalized := strings.ToLower(strings.TrimSpace(clientLocale))
	normalized = strings.ReplaceAll(normalized, "_", "-")
	return normalized
}

func getAIAssistantPreferredReplyLanguage(clientLocale string) string {
	normalized := normalizeAIAssistantClientLocale(clientLocale)

	switch {
	case strings.HasPrefix(normalized, "zh-hant"), strings.HasPrefix(normalized, "zh-tw"), strings.HasPrefix(normalized, "zh-hk"), strings.HasPrefix(normalized, "zh-mo"):
		return "Traditional Chinese"
	case strings.HasPrefix(normalized, "zh"):
		return "Simplified Chinese"
	case strings.HasPrefix(normalized, "ja"):
		return "Japanese"
	case strings.HasPrefix(normalized, "ko"):
		return "Korean"
	case strings.HasPrefix(normalized, "fr"):
		return "French"
	case strings.HasPrefix(normalized, "de"):
		return "German"
	case strings.HasPrefix(normalized, "es"):
		return "Spanish"
	case strings.HasPrefix(normalized, "it"):
		return "Italian"
	case strings.HasPrefix(normalized, "ru"):
		return "Russian"
	case strings.HasPrefix(normalized, "uk"):
		return "Ukrainian"
	case strings.HasPrefix(normalized, "vi"):
		return "Vietnamese"
	case strings.HasPrefix(normalized, "th"):
		return "Thai"
	case strings.HasPrefix(normalized, "pt-br"):
		return "Brazilian Portuguese"
	case strings.HasPrefix(normalized, "pt"):
		return "Portuguese"
	case strings.HasPrefix(normalized, "tr"):
		return "Turkish"
	case strings.HasPrefix(normalized, "nl"):
		return "Dutch"
	case strings.HasPrefix(normalized, "sl"):
		return "Slovenian"
	case strings.HasPrefix(normalized, "ta"):
		return "Tamil"
	case strings.HasPrefix(normalized, "kn"):
		return "Kannada"
	default:
		return "English"
	}
}
