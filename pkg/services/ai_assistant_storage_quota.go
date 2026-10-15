package services

import (
	"encoding/json"

	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ReserveConversation reserves the full bounded header, including future summary
// growth. Header settlement converts reserved bytes to used bytes, not more space.
func (tx *AssistantStorageTx) ReserveConversation(conversation *models.AssistantConversation, limits *models.AssistantLimits) error {
	n := int64(limits.ConversationHeaderBytes)
	if conversation.Uid != tx.State.Uid || n <= 0 || tx.State.ConversationCount >= limits.Conversations || !assistantQuotaFits(tx.State.UsedBytes, tx.State.ReservedBytes, n, int64(limits.UserBytes)) || n > int64(limits.ConversationBytes) {
		return models.AssistantReasonQuotaExceeded
	}
	tx.State.ConversationCount++
	tx.State.ReservedBytes += n
	conversation.HeaderReservedBytes = n
	conversation.ReservedBytes = n
	return nil
}

func (tx *AssistantStorageTx) ReserveRun(conversation *models.AssistantConversation, run *models.AssistantMessage, limits *models.AssistantLimits) error {
	n := int64(limits.AttemptBytes)
	if conversation.Uid != tx.State.Uid || run.Uid != tx.State.Uid || run.ConversationId != conversation.Id || run.ReservedBytes != 0 || run.ReservationSettled || n <= 0 {
		return models.AssistantReasonConversationConflict
	}
	if tx.State.QueuedCount >= limits.QueuedPerUser {
		return models.AssistantReasonQueueFull
	}
	if !assistantQuotaFits(tx.State.UsedBytes, tx.State.ReservedBytes, n, int64(limits.UserBytes)) || !assistantQuotaFits(conversation.UsedBytes, conversation.ReservedBytes, n, int64(limits.ConversationBytes)) {
		return models.AssistantReasonQuotaExceeded
	}
	raw, err := json.Marshal(limits)
	if err != nil {
		return err
	}
	run.LimitsJSON = string(raw)
	run.ReservedBytes = n
	run.QuotaState = "queued"
	tx.State.QueuedCount++
	tx.State.ReservedBytes += n
	conversation.ReservedBytes += n
	return nil
}

// MarkRunRunning transfers the reservation's execution slot on claim. Callers
// persist the run and conversation in the same transaction as this operation.
func (tx *AssistantStorageTx) MarkRunRunning(run *models.AssistantMessage, limits *models.AssistantLimits) error {
	if run.Uid != tx.State.Uid || run.ReservationSettled || run.QuotaState != "queued" || tx.State.QueuedCount < 1 {
		return models.AssistantReasonConversationConflict
	}
	if tx.State.RunningCount >= limits.RunningPerUser {
		return models.AssistantReasonQueueFull
	}
	tx.State.QueuedCount--
	tx.State.RunningCount++
	run.QuotaState = "running"
	return nil
}

// SettleRun uses the persisted flag of a freshly loaded run. It is safe to call
// twice; all model updates must be saved within this same user transaction.
func (tx *AssistantStorageTx) SettleRun(conversation *models.AssistantConversation, run *models.AssistantMessage, actualBytes int64) error {
	if run.Uid != tx.State.Uid || conversation.Uid != tx.State.Uid || run.ConversationId != conversation.Id {
		return models.AssistantReasonNotAccessible
	}
	if run.ReservationSettled {
		return nil
	}
	if actualBytes < 0 || actualBytes > run.ReservedBytes || run.UsedBytes > actualBytes || run.ReservedBytes > conversation.ReservedBytes || run.ReservedBytes > tx.State.ReservedBytes {
		return models.AssistantReasonQuotaExceeded
	}
	delta := actualBytes - run.UsedBytes
	tx.State.ReservedBytes -= run.ReservedBytes
	conversation.ReservedBytes -= run.ReservedBytes
	tx.State.UsedBytes += delta
	conversation.UsedBytes += delta
	if run.QuotaState == "queued" {
		if tx.State.QueuedCount < 1 {
			return models.AssistantReasonStorageUnavailable
		}
		tx.State.QueuedCount--
	} else if run.QuotaState == "running" {
		if tx.State.RunningCount < 1 {
			return models.AssistantReasonStorageUnavailable
		}
		tx.State.RunningCount--
	} else {
		return models.AssistantReasonStorageUnavailable
	}
	run.UsedBytes = actualBytes
	run.ReservedBytes = 0
	run.ReservationSettled = true
	run.QuotaState = ""
	return nil
}

// SettleHeader keeps the unused part of its original reservation for future
// summary growth. Header shrinkage returns capacity to that same reservation.
func (tx *AssistantStorageTx) SettleHeader(conversation *models.AssistantConversation, actualBytes int64) error {
	if conversation.Uid != tx.State.Uid {
		return models.AssistantReasonNotAccessible
	}
	delta := actualBytes - conversation.HeaderUsedBytes
	if actualBytes < 0 || delta > conversation.HeaderReservedBytes || delta > tx.State.ReservedBytes {
		return models.AssistantReasonQuotaExceeded
	}
	conversation.HeaderUsedBytes = actualBytes
	conversation.HeaderReservedBytes -= delta
	conversation.UsedBytes += delta
	conversation.ReservedBytes -= delta
	tx.State.UsedBytes += delta
	tx.State.ReservedBytes -= delta
	return nil
}

func assistantQuotaFits(used, reserved, additional, limit int64) bool {
	return used >= 0 && reserved >= 0 && additional >= 0 && limit >= used && reserved <= limit-used && additional <= limit-used-reserved
}

// AssistantStorageBytes counts a fixed public/storage projection, including JSON
// escaping and metadata. Do not pass raw provider responses or whole SQL rows.
func AssistantStorageBytes(projection any) (int64, error) {
	raw, err := json.Marshal(projection)
	return int64(len(raw)), err
}
