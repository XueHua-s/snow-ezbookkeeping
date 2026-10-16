package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// ClearAssistantData revokes old commands and workers before reclaiming content.
// The user state and minimum, body-free receipts deliberately survive a clear.
func (tx *AssistantStorageTx) ClearAssistantData() error {
	tx.State.DeletionEpoch++
	tx.State.MemoryGenerationEpoch++
	tx.State.MemoryUseEpoch++
	// A clear invalidates every source, so no per-conversation cascade or quota
	// subtraction is needed. Fencing, terminal runs and zeroed counters commit
	// atomically; a late worker cannot settle the same reservation again.
	if _, err := tx.Session.Where("uid = ? AND role = ?", tx.State.Uid, "assistant").In("run_state", []models.AssistantRunState{models.AssistantRunQueued, models.AssistantRunPreparing, models.AssistantRunRunning, models.AssistantRunCancelRequested}).SetExpr("snapshot_revision", "snapshot_revision + 1").Cols("run_state", "cancel_requested", "completed_at", "reservation_settled", "reserved_bytes", "used_bytes", "quota_state").Update(&models.AssistantMessage{RunState: models.AssistantRunCanceled, CancelRequested: true, CompletedAt: tx.Now.UnixMilli(), ReservationSettled: true, QuotaState: ""}); err != nil {
		return err
	}
	if _, err := tx.Session.Where("uid = ? AND deleted = ?", tx.State.Uid, false).SetExpr("metadata_version", "metadata_version + 1").SetExpr("context_version", "context_version + 1").SetExpr("source_version", "source_version + 1").Cols("deleted", "deleted_at", "active_run_id", "used_bytes", "reserved_bytes", "header_used_bytes", "header_reserved_bytes").Update(&models.AssistantConversation{Deleted: true, DeletedAt: tx.Now.UnixMilli()}); err != nil {
		return err
	}

	if _, err := tx.Session.Where("uid = ?", tx.State.Uid).Delete(&models.AssistantMemory{}); err != nil {
		return err
	}
	tx.State.UsedBytes = 0
	tx.State.ReservedBytes = 0
	tx.State.ConversationCount = 0
	tx.State.RunningCount = 0
	tx.State.QueuedCount = 0
	tx.State.ActiveMemoryCount = 0
	tx.State.CandidateMemoryCount = 0
	if err := tx.TombstoneReceipts(0, 0, true); err != nil {
		return err
	}
	return nil
}

// InvalidateLedgerContext keeps the readable conversation while fencing every
// in-flight financial answer and discarding derived context from the old ledger.
func (s *AssistantStorageService) InvalidateLedgerContext(ctx core.Context, uid int64) error {
	return s.WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		return tx.invalidateLedgerContext()
	})
}

// invalidateLedgerContext also runs inside the transaction that deletes ledger
// rows, so a run accepted after an earlier fence cannot survive that commit.
func (tx *AssistantStorageTx) invalidateLedgerContext() error {
	var after int64
	for {
		var rows []models.AssistantConversation
		if err := tx.Session.Where("uid = ? AND deleted = ? AND id > ?", tx.State.Uid, false, after).Asc("id").Limit(50).Find(&rows); err != nil {
			return err
		}
		if len(rows) == 0 {
			break
		}
		for i := range rows {
			c := &rows[i]
			after = c.Id
			if err := tx.CancelConversationRuns(c.Id, models.AssistantReasonContextChanged); err != nil {
				return err
			}
			var err error
			c, err = assistantConversationLoad(tx.Session, tx.State.Uid, c.Id)
			if err != nil {
				return err
			}
			c.ContextVersion++
			c.SourceVersion++
			c.ContextInvalidatedThroughTurn = c.NextTurnSequence - 1
			c.SourceInvalidated = true
			c.Summary = ""
			c.SummaryThrough = 0
			c.SummarySourcesJSON = ""
			c.SeedContext = ""
			c.SeedSourcesJSON = ""
			c.FrozenPrefixJSON = ""
			if err = assistantSettleConversationHeader(tx, c); err != nil {
				return err
			}
			if err = assistantConversationSave(tx, c); err != nil {
				return err
			}
		}
	}
	tx.State.MemoryGenerationEpoch++
	return nil
}
