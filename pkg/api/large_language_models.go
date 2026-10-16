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
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/mayswind/ezbookkeeping/pkg/templates"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"strings"
	"time"
)

// LargeLanguageModelsApi represents large language models api
type LargeLanguageModelsApi struct {
	ApiUsingConfig
	assistant             *assistantDependencies
	experience            *assistantExperienceDependencies
	transactions          *services.TransactionService
	transactionCategories *services.TransactionCategoryService
	transactionTags       *services.TransactionTagService
	accounts              *services.AccountService
	users                 *services.UserService
}

// Initialize a large language models api singleton instance
var (
	LargeLanguageModels = &LargeLanguageModelsApi{
		ApiUsingConfig: ApiUsingConfig{
			container: settings.Container,
		},
		transactions:          services.Transactions,
		transactionCategories: services.TransactionCategories,
		transactionTags:       services.TransactionTags,
		accounts:              services.Accounts,
		users:                 services.Users,
	}
)

// RecognizeTransactionTextHandler returns the recognized transaction text result
func (a *LargeLanguageModelsApi) RecognizeTransactionTextHandler(c *core.WebContext) (any, *errs.Error) {
	if a.CurrentConfig().TextRecognitionLLMConfig == nil || a.CurrentConfig().TextRecognitionLLMConfig.LLMProvider == "" || !a.CurrentConfig().TransactionFromAITextRecognition {
		return nil, errs.ErrLargeLanguageModelProviderNotEnabled
	}

	clientTimezone, err := c.GetClientTimezone()

	if err != nil {
		log.Warnf(c, "[large_language_models.RecognizeTransactionTextHandler] cannot get client timezone, because %s", err.Error())
		return nil, errs.ErrClientTimezoneOffsetInvalid
	}

	uid := c.GetCurrentUid()
	user, err := a.users.GetUserById(c, uid)

	if err != nil {
		if !errs.IsCustomError(err) {
			log.Warnf(c, "[large_language_models.RecognizeTransactionTextHandler] failed to get user for user \"uid:%d\", because %s", uid, err.Error())
		}

		return false, errs.ErrUserNotFound
	}

	if user.FeatureRestriction.Contains(core.USER_FEATURE_RESTRICTION_TYPE_CREATE_TRANSACTION_FROM_AI_TEXT_RECOGNITION) {
		return false, errs.ErrNotPermittedToPerformThisAction
	}

	var textRecognitionReq models.TransactionTextRecognitionRequest
	err = c.ShouldBindJSON(&textRecognitionReq)

	if err != nil {
		log.Warnf(c, "[large_language_models.RecognizeTransactionTextHandler] parse request failed, because %s", err.Error())
		return nil, errs.NewIncompleteOrIncorrectSubmissionError(err)
	}

	if len(textRecognitionReq.Text) == 0 {
		log.Warnf(c, "[large_language_models.RecognizeTransactionTextHandler] there is no text in request for user \"uid:%d\"", uid)
		return nil, errs.ErrNoAIRecognitionText
	}

	text := strings.TrimSpace(textRecognitionReq.Text)

	if len(text) == 0 {
		log.Warnf(c, "[large_language_models.RecognizeTransactionTextHandler] the text in request is empty for user \"uid:%d\"", uid)
		return nil, errs.ErrAIRecognitionTextIsEmpty
	}

	accountNames, accountMap, incomeCategoryNames, expenseCategoryNames, transferCategoryNames, incomeCategoryMap, expenseCategoryMap, transferCategoryMap, tagNames, tagMap, err := a.getUserEssentialData(c, uid)

	if err != nil {
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	systemPrompt, err := templates.GetTemplate(templates.SYSTEM_PROMPT_TRANSACTION_TEXT_RECOGNITION)

	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeTransactionTextHandler] failed to get system prompt template for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	systemPromptParams := map[string]any{
		"CurrentDateTime":          utils.FormatUnixTimeToLongDateTime(time.Now().Unix(), clientTimezone),
		"AllExpenseCategoryNames":  strings.Join(expenseCategoryNames, "\n"),
		"AllIncomeCategoryNames":   strings.Join(incomeCategoryNames, "\n"),
		"AllTransferCategoryNames": strings.Join(transferCategoryNames, "\n"),
		"AllAccountNames":          strings.Join(accountNames, "\n"),
		"AllTagNames":              strings.Join(tagNames, "\n"),
	}

	var bodyBuffer bytes.Buffer
	err = systemPrompt.Execute(&bodyBuffer, systemPromptParams)

	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeTransactionTextHandler] failed to get final system prompt from template for user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	llmRequest := &data.LargeLanguageModelRequest{
		Stream:         false,
		SystemPrompt:   strings.ReplaceAll(bodyBuffer.String(), "\r\n", "\n"),
		UserPrompt:     []byte(text),
		UserPromptType: data.LARGE_LANGUAGE_MODEL_REQUEST_PROMPT_TYPE_TEXT,
	}

	llmResponse, err := llm.Container.GetJsonResponseByTextRecognitionModel(c, c.GetCurrentUid(), llmRequest)

	if err != nil {
		log.Errorf(c, "[large_language_models.RecognizeTransactionTextHandler] failed to get llm response user \"uid:%d\", because %s", uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	if llmResponse == nil || len(llmResponse.Content) == 0 || strings.HasPrefix(llmResponse.Content, "{}") {
		return nil, errs.ErrNoTransactionInformation
	}

	var result *models.RecognizedTransactionResult

	if err := json.Unmarshal([]byte(llmResponse.Content), &result); err != nil {
		log.Errorf(c, "[large_language_models.RecognizeTransactionTextHandler] failed to unmarshal recognized transaction text result from llm response \"%s\" for user \"uid:%d\", because %s", llmResponse.Content, uid, err.Error())
		return nil, errs.Or(err, errs.ErrOperationFailed)
	}

	return a.parseRecognizedTransactionResponse(c, uid, clientTimezone, result, accountMap, expenseCategoryMap, incomeCategoryMap, transferCategoryMap, tagMap)
}
