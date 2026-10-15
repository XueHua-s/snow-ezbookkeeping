package api

import (
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// AssistantExperienceError maps only controlled reasons to the public envelope.
// Database and provider details must never cross the HTTP boundary.
func assistantExperienceError(err error) *errs.Error {
	if err == nil {
		return nil
	}
	var reason models.AssistantReasonCode
	if !errors.As(err, &reason) {
		reason = models.AssistantReasonStorageUnavailable
	}
	// Indices are part of the public error contract; append new reasons only.
	reasons := []models.AssistantReasonCode{models.AssistantReasonInputTooLong, models.AssistantReasonConversationConflict, models.AssistantReasonDataEpochConflict, models.AssistantReasonMemoryConflict, models.AssistantReasonContextChanged, models.AssistantReasonAlreadyRunning, models.AssistantReasonQuotaExceeded, models.AssistantReasonQueueFull, models.AssistantReasonSubscriptionLimit, models.AssistantReasonCursorExpired, models.AssistantReasonContextLimit, models.AssistantReasonOutputLimit, models.AssistantReasonEmptyResponse, models.AssistantReasonStorageUnavailable, models.AssistantReasonProviderUnavailable, models.AssistantReasonInterrupted, models.AssistantReasonNotAccessible, models.AssistantReasonQueueTimeout, models.AssistantReasonRunTimeout, models.AssistantReasonExecutionLimit, models.AssistantReasonGenerationUnavailable}
	for i, known := range reasons {
		if reason == known {
			return errs.NewErrorWithContext(errs.NewNormalError(errs.NormalSubcategoryLargeLanguageModel, int32(60+i), reason.HTTPStatus(), string(reason)), map[string]string{"reasonCode": string(reason)})
		}
	}
	return assistantExperienceError(models.AssistantReasonStorageUnavailable)
}
