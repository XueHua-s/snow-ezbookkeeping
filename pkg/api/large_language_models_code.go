package api

import (
	"bytes"
	"encoding/json/v2"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/templates"
	"html/template"
	"strings"
)

// GenerateCodeHandler generates complete code for a supported coding task
func (a *LargeLanguageModelsApi) GenerateCodeHandler(c *core.WebContext) (any, *errs.Error) {
	if a.CurrentConfig().CodingAssistantLLMConfig == nil || a.CurrentConfig().CodingAssistantLLMConfig.LLMProvider == "" {
		return nil, errs.ErrLargeLanguageModelProviderNotEnabled
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[large_language_models.GenerateCodeHandler] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}

		return nil, errs.ErrUserNotFound
	}

	var codingAssistantReq models.CodingAssistantRequest
	err = c.ShouldBindJSON(&codingAssistantReq)

	if err != nil {
		log.Warnf(c, "[large_language_models.GenerateCodeHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	var systemPromptTemplateName templates.KnownTemplate
	systemPromptParams := map[string]any{}

	if codingAssistantReq.Type == models.CODING_ASSISTANT_TYPE_INSIGHTS_EXPLORER_CUSTOM_CHART {
		if !a.CurrentConfig().InsightsExplorerCodingAssistant {
			return nil, errs.ErrLargeLanguageModelProviderNotEnabled
		}

		if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_GENERATE_CUSTOM_CHART_SCRIPTS_BY_AI_CODING_ASSISTANT) {
			return nil, errs.ErrNotPermittedToPerformThisAction
		}

		systemPromptTemplateName = templates.SYSTEM_PROMPT_INSIGHTS_EXPLORER_CUSTOM_CHART
		systemPromptParams["CurrentCode"] = template.HTML(codingAssistantReq.Code)
	} else {
		return nil, errs.ErrParameterInvalid
	}

	userPrompt := strings.TrimSpace(codingAssistantReq.UserPrompt)

	if len(userPrompt) == 0 {
		log.Warnf(c, "[large_language_models.GenerateCodeHandler] the text in user prompt is empty for user \"uid:%d\"", uid)
		return nil, errs.ErrNoAICodingPrompt
	}

	systemPrompt, err := templates.GetTemplate(systemPromptTemplateName)

	if err != nil {
		log.Errorf(c, "[large_language_models.GenerateCodeHandler] failed to get coding prompt template, because %s", err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	var bodyBuffer bytes.Buffer
	err = systemPrompt.Execute(&bodyBuffer, systemPromptParams)

	if err != nil {
		log.Errorf(c, "[large_language_models.GenerateCodeHandler] failed to get final system prompt from template for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	llmRequest := &data.LargeLanguageModelRequest{
		Stream:         false,
		SystemPrompt:   strings.ReplaceAll(bodyBuffer.String(), "\r\n", "\n"),
		UserPrompt:     []byte(userPrompt),
		UserPromptType: data.LARGE_LANGUAGE_MODEL_REQUEST_PROMPT_TYPE_TEXT,
	}

	llmResponse, err := llm.Container.GetJsonResponseByCodingAssistantModel(c, c.GetCurrentUid(), llmRequest)

	if err != nil {
		log.Errorf(c, "[large_language_models.GenerateCodeHandler] failed to get llm response for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if llmResponse == nil || len(llmResponse.Content) == 0 || strings.HasPrefix(llmResponse.Content, "{}") {
		return nil, errs.ErrNoAICodingResult
	}

	var result *models.CodingAssistantResult

	if err := json.Unmarshal([]byte(llmResponse.Content), &result); err != nil {
		log.Errorf(c, "[large_language_models.GenerateCodeHandler] failed to parse coding response for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.ErrNoAICodingResult
	}

	if result == nil {
		return nil, errs.ErrNoAICodingResult
	}

	code := strings.TrimSpace(result.Code)
	failureReason := strings.TrimSpace(result.FailureReason)

	if (len(code) == 0 && len(failureReason) == 0) || (len(code) > 0 && len(failureReason) > 0) {
		log.Warnf(c, "[large_language_models.GenerateCodeHandler] the response of coding assistant is invalid for user \"uid:%d\"", uid)
		return nil, errs.ErrNoAICodingResult
	}

	return &models.CodingAssistantResponse{
		Code:          code,
		FailureReason: failureReason,
	}, nil
}
