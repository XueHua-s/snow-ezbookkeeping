package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
)

type assistantExperienceDependencies struct {
	limiter             assistantRequestLimiter
	lifecycleMu         sync.Mutex
	backgroundCancel    context.CancelFunc
	backgroundDone      chan struct{}
	cleanup             *services.AssistantCleanupService
	storage             *services.AssistantStorageService
	conversations       *services.AssistantConversationService
	memories            *services.AssistantMemoryService
	runs                *services.AssistantRunService
	generationAvailable func(core.Context, int64) bool
}

func (a *LargeLanguageModelsApi) assistantExperienceReady(c *core.WebContext) *errs.Error {
	c.Header("Cache-Control", "no-store")
	if c.GetCurrentUid() <= 0 {
		return errs.ErrUserNotFound
	}
	if a.experience == nil {
		return assistantExperienceError(models.AssistantReasonStorageUnavailable)
	}
	return nil
}

// Decode before mutation: limit actual bytes (including chunked requests), reject
// unknown fields and additional JSON values, and never log rejected content.
func (a *LargeLanguageModelsApi) assistantDecode(c *core.WebContext, value any) *errs.Error {
	if err := a.assistantExperienceReady(c); err != nil {
		return err
	}
	limit := int64(a.CurrentConfig().GetAIAssistantExperienceConfig().Limits.RequestBytes)
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, limit+1))
	if int64(len(body)) > limit || c.Request.ContentLength > limit {
		return errs.ErrAssistantRequestTooLarge
	}
	if err != nil {
		return errs.ErrIncompleteOrIncorrectSubmission
	}
	if strings.TrimSpace(string(body)) == "null" {
		return errs.ErrIncompleteOrIncorrectSubmission
	}
	if err := json.Unmarshal(body, value, json.RejectUnknownMembers(true)); err != nil {
		return errs.ErrIncompleteOrIncorrectSubmission
	}

	return nil
}
func assistantQuery(c *core.WebContext, allowed ...string) *errs.Error {
	query := c.Request.URL.Query()
	for key, values := range query {
		found := false
		for _, name := range allowed {
			if key == name {
				found = true
				break
			}
		}
		if !found || len(values) != 1 || len(values[0]) > 8192 {
			return errs.ErrIncompleteOrIncorrectSubmission
		}
	}
	return nil
}
func assistantQueryInt(c *core.WebContext, name string, max int64) (int64, *errs.Error) {
	raw := c.Query(name)
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value < 0 || value > max {
		return 0, errs.ErrIncompleteOrIncorrectSubmission
	}
	return value, nil
}
func assistantQueryBool(c *core.WebContext, name string) (bool, *errs.Error) {
	raw := c.Query(name)
	if raw == "" {
		return false, nil
	}
	if raw == "true" {
		return true, nil
	}
	if raw == "false" {
		return false, nil
	}
	return false, errs.ErrIncompleteOrIncorrectSubmission
}
func (a *LargeLanguageModelsApi) AssistantCapabilitiesHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	release, e := a.assistantLimit(c, assistantStatusRequest, false)
	if e != nil {
		return nil, e
	}
	defer release()

	if err := assistantQuery(c); err != nil {
		return nil, err
	}
	state, err := a.experience.storage.GetUserState(assistantHTTPContext(c), c.GetCurrentUid())
	if err != nil {
		return nil, assistantExperienceError(err)
	}
	mac := hmac.New(sha256.New, []byte(a.CurrentConfig().SecretKey))
	mac.Write([]byte("assistant-owner-v1:" + a.CurrentConfig().RootUrl + ":" + strconv.FormatInt(c.GetCurrentUid(), 10)))
	result := models.AssistantCapabilities{SchemaVersion: models.AssistantSchemaVersion, HistoryAvailable: true, MemoryAvailable: true, OwnerScope: hex.EncodeToString(mac.Sum(nil)), ServerTime: time.Now().UTC().Format(time.RFC3339Nano), DeletionEpoch: state.DeletionEpoch, SettingsVersion: state.SettingsVersion, MemoryUseEpoch: state.MemoryUseEpoch, MemoryEnabled: state.MemoryEnabled, Limits: a.CurrentConfig().GetAIAssistantExperienceConfig().Limits, Usage: models.AssistantUsage{UsedBytes: state.UsedBytes, ReservedBytes: state.ReservedBytes, ConversationCount: state.ConversationCount, RunningCount: state.RunningCount, QueuedCount: state.QueuedCount, ActiveMemoryCount: state.ActiveMemoryCount, CandidateMemoryCount: state.CandidateMemoryCount}}
	result.GenerationAvailable = a.experience.generationAvailable != nil && a.experience.generationAvailable(assistantHTTPContext(c), c.GetCurrentUid())
	if !result.GenerationAvailable {
		result.UnavailableReason = models.AssistantReasonProviderUnavailable
	}
	return result, nil
}

func (a *LargeLanguageModelsApi) AssistantConversationSearchHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantSearchRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	release, e := a.assistantLimit(c, assistantSearchExportRequest, false)
	if e != nil {
		return nil, e
	}
	defer release()

	result, err := a.experience.conversations.Search(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantConversationUpdateHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantUpdateConversationRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.conversations.Update(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantConversationDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantDeleteConversationRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.conversations.Delete(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantConversationClearHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantClearRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.conversations.Clear(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantConversationBranchHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantBranchRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.conversations.Branch(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantMessageSendHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantSendRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	release, e := a.assistantLimit(c, assistantGenerationRequest, false)
	if e != nil {
		return nil, e
	}
	defer release()

	result, err := a.experience.conversations.Send(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantMessageRetryHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantRetryRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	release, e := a.assistantLimit(c, assistantGenerationRequest, false)
	if e != nil {
		return nil, e
	}
	defer release()

	result, err := a.experience.conversations.Retry(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantMessageSelectVersionHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantSelectVersionRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.conversations.SelectVersion(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantRunCancelHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantCancelRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.runs.Cancel(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantMemorySaveHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantSaveMemoryRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.memories.Save(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantMemoryDeleteHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantDeleteMemoryRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.memories.Delete(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantMemorySettingsHandler(c *core.WebContext) (any, *errs.Error) {
	var request models.AssistantMemorySettingsRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return nil, err
	}
	result, err := a.experience.memories.Settings(assistantHTTPContext(c), c.GetCurrentUid(), request)
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantCommandGetHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	release, e := a.assistantLimit(c, assistantStatusRequest, false)
	if e != nil {
		return nil, e
	}
	defer release()

	if err := assistantQuery(c, "clientRequestId"); err != nil {
		return nil, err
	}
	if c.Query("clientRequestId") == "" {
		return nil, errs.ErrIncompleteOrIncorrectSubmission
	}
	result, err := a.experience.storage.GetCommand(assistantHTTPContext(c), c.GetCurrentUid(), c.Query("clientRequestId"))
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantConversationStateHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	release, e := a.assistantLimit(c, assistantStatusRequest, false)
	if e != nil {
		return nil, e
	}
	defer release()

	if err := assistantQuery(c, "id"); err != nil {
		return nil, err
	}
	if c.Query("id") == "" {
		return nil, errs.ErrIncompleteOrIncorrectSubmission
	}
	result, err := a.experience.conversations.State(assistantHTTPContext(c), c.GetCurrentUid(), c.Query("id"))
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantMessageGetHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	if err := assistantQuery(c, "assistantMessageId"); err != nil {
		return nil, err
	}
	if c.Query("assistantMessageId") == "" {
		return nil, errs.ErrIncompleteOrIncorrectSubmission
	}
	result, err := a.experience.conversations.GetMessage(assistantHTTPContext(c), c.GetCurrentUid(), c.Query("assistantMessageId"))
	return result, assistantExperienceError(err)
}

func (a *LargeLanguageModelsApi) AssistantConversationListHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	if err := assistantQuery(c, "cursor", "limit", "archived"); err != nil {
		return nil, err
	}
	limit, e := assistantQueryInt(c, "limit", int64(a.CurrentConfig().GetAIAssistantExperienceConfig().Limits.PageSizeMax))
	if e != nil {
		return nil, e
	}
	archived, e := assistantQueryBool(c, "archived")
	if e != nil {
		return nil, e
	}
	result, err := a.experience.conversations.List(assistantHTTPContext(c), c.GetCurrentUid(), models.AssistantListRequest{Cursor: c.Query("cursor"), Limit: int(limit), Archived: archived})
	return result, assistantExperienceError(err)
}
func (a *LargeLanguageModelsApi) AssistantConversationGetHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	if err := assistantQuery(c, "id", "beforeTurnSequence", "aroundMessageId", "limit"); err != nil {
		return nil, err
	}
	limit, e := assistantQueryInt(c, "limit", int64(a.CurrentConfig().GetAIAssistantExperienceConfig().Limits.PageSizeMax))
	if e != nil {
		return nil, e
	}
	before, e := assistantQueryInt(c, "beforeTurnSequence", 1<<63-1)
	if e != nil {
		return nil, e
	}
	if c.Query("id") == "" || before > 0 && c.Query("aroundMessageId") != "" {
		return nil, errs.ErrIncompleteOrIncorrectSubmission
	}
	result, err := a.experience.conversations.Get(assistantHTTPContext(c), c.GetCurrentUid(), models.AssistantGetConversationRequest{Id: c.Query("id"), BeforeTurnSequence: before, AroundMessageId: c.Query("aroundMessageId"), Limit: int(limit)})
	return result, assistantExperienceError(err)
}
func (a *LargeLanguageModelsApi) AssistantMessageVersionsHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	if err := assistantQuery(c, "userMessageId", "cursor", "limit"); err != nil {
		return nil, err
	}
	limit, e := assistantQueryInt(c, "limit", int64(a.CurrentConfig().GetAIAssistantExperienceConfig().Limits.PageSizeMax))
	if e != nil {
		return nil, e
	}
	if c.Query("userMessageId") == "" {
		return nil, errs.ErrIncompleteOrIncorrectSubmission
	}
	result, err := a.experience.conversations.Versions(assistantHTTPContext(c), c.GetCurrentUid(), models.AssistantVersionsRequest{UserMessageId: c.Query("userMessageId"), Cursor: c.Query("cursor"), Limit: int(limit)})
	return result, assistantExperienceError(err)
}
func (a *LargeLanguageModelsApi) AssistantRunGetHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	release, e := a.assistantLimit(c, assistantStatusRequest, false)
	if e != nil {
		return nil, e
	}
	defer release()

	if err := assistantQuery(c, "runId", "clientRequestId"); err != nil {
		return nil, err
	}
	runID := c.Query("runId")
	requestID := c.Query("clientRequestId")
	if (runID == "") == (requestID == "") {
		return nil, errs.ErrIncompleteOrIncorrectSubmission
	}
	if requestID != "" {
		receipt, err := a.experience.storage.GetCommand(assistantHTTPContext(c), c.GetCurrentUid(), requestID)
		if err != nil {
			return nil, assistantExperienceError(err)
		}
		if receipt.Status != "accepted" || receipt.RunId == "" {
			return nil, assistantExperienceError(models.AssistantReasonNotAccessible)
		}
		runID = receipt.RunId
	}
	result, err := a.experience.runs.GetRunSnapshot(assistantHTTPContext(c), c.GetCurrentUid(), runID)
	return result, assistantExperienceError(err)
}
func (a *LargeLanguageModelsApi) AssistantMemoryListHandler(c *core.WebContext) (any, *errs.Error) {
	if err := a.assistantExperienceReady(c); err != nil {
		return nil, err
	}
	if err := assistantQuery(c); err != nil {
		return nil, err
	}
	result, err := a.experience.memories.List(assistantHTTPContext(c), c.GetCurrentUid())
	return result, assistantExperienceError(err)
}

// Export is deliberately a raw response handler: after the first byte a second
// JSON envelope must never be appended. The service checks ownership before writing.
func (a *LargeLanguageModelsApi) AssistantConversationExportHandler(c *core.WebContext) *errs.Error {
	var request models.AssistantExportRequest
	if err := a.assistantDecode(c, &request); err != nil {
		return err
	}
	release, e := a.assistantLimit(c, assistantSearchExportRequest, true)
	if e != nil {
		return e
	}
	defer release()

	if request.All == (request.Id != "") || (request.Format != "json" && request.Format != "markdown") {
		return errs.ErrIncompleteOrIncorrectSubmission
	}
	controller := http.NewResponseController(c.Writer)
	if err := controller.SetWriteDeadline(time.Now().Add(time.Duration(a.CurrentConfig().GetAIAssistantExperienceConfig().Limits.ExportTimeoutSeconds) * time.Second)); err != nil {
		return assistantExperienceError(models.AssistantReasonStorageUnavailable)
	}
	defer controller.SetWriteDeadline(time.Time{})
	extension, contentType := "json", "application/json; charset=utf-8"
	if request.Format == "markdown" {
		extension = "md"
		contentType = "text/markdown; charset=utf-8"
	}
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", `attachment; filename="assistant-export.`+extension+`"`)
	c.Header("X-Content-Type-Options", "nosniff")
	err := a.experience.conversations.Export(assistantHTTPContext(c), c.GetCurrentUid(), request, c.Writer)
	if err != nil && !c.Writer.Written() {
		c.Header("Content-Disposition", "")
		c.Header("Content-Type", "application/json; charset=utf-8")
		return assistantExperienceError(err)
	}
	return nil
}

func assistantHTTPContext(c *core.WebContext) core.Context {
	return core.NewAssistantContext(c.Request.Context(), c.GetContextId(), c.GetClientLocale(), c.ClientIP())
}
