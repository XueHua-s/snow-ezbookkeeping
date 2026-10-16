package services

import (
	"strconv"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ValidateDependencies is used before starting upstream work and committing
// checkpoints. An unrelated memory update does not invalidate the frozen input.
func (tx *AssistantStorageTx) ValidateDependencies(conversation *models.AssistantConversation, dependencies *models.AssistantInputDependencies, acceptedLimits ...models.AssistantLimits) error {
	if tx.State.AccountUnavailable || conversation.Uid != tx.State.Uid || conversation.Deleted || dependencies.DeletionEpoch != tx.State.DeletionEpoch {
		return models.AssistantReasonContextChanged
	}
	limits := tx.Limits
	if len(acceptedLimits) > 0 {
		limits = acceptedLimits[0]
	}
	sourceBytes, err := AssistantStorageBytes(dependencies.Sources)
	if err != nil {
		return models.AssistantReasonContextLimit
	}
	if sourceBytes > int64(limits.ConversationHeaderBytes) || len(dependencies.Memories) > limits.Memories {
		return models.AssistantReasonContextLimit
	}
	if len(dependencies.Memories) > 0 && (!tx.State.MemoryEnabled || conversation.MemoryMode == "off" || dependencies.MemoryMode == "off") {
		return models.AssistantReasonContextChanged
	}
	for _, source := range dependencies.Sources {
		id, err := strconv.ParseInt(source.ConversationId, 10, 64)
		if err != nil || id <= 0 {
			return models.AssistantReasonContextChanged
		}
		sourceConversation := &models.AssistantConversation{}
		found, err := tx.Session.Where("uid = ? AND id = ? AND deleted = ?", tx.State.Uid, id, false).Get(sourceConversation)
		if err != nil {
			return err
		}
		if !found || sourceConversation.SourceVersion != source.Version {
			return models.AssistantReasonContextChanged
		}
		if source.MessageId != "" {
			messageId, err := strconv.ParseInt(source.MessageId, 10, 64)
			if err != nil || messageId <= 0 {
				return models.AssistantReasonContextChanged
			}
			found, err := tx.Session.Where("uid = ? AND conversation_id = ? AND id = ?", tx.State.Uid, id, messageId).Exist(&models.AssistantMessage{})
			if err != nil {
				return err
			}
			if !found {
				return models.AssistantReasonContextChanged
			}
		}
	}
	for _, used := range dependencies.Memories {
		id, err := strconv.ParseInt(used.Id, 10, 64)
		if err != nil || id <= 0 {
			return models.AssistantReasonContextChanged
		}
		memory := &models.AssistantMemory{}
		found, err := tx.Session.Where("uid = ? AND id = ? AND excluded = ?", tx.State.Uid, id, false).Get(memory)
		if err != nil {
			return err
		}
		if !found || memory.Version != used.Version || memory.Status != "active" || (memory.ExpiresAt > 0 && memory.ExpiresAt <= tx.Now.UnixMilli()) {
			return models.AssistantReasonContextChanged
		}
	}
	dependencies.MemoryUseEpoch = tx.State.MemoryUseEpoch
	return nil
}
