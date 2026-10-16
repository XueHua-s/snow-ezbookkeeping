package services

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func (s *AssistantRunService) scan(parent context.Context, work *sync.WaitGroup) {
	ctx := core.NewAssistantContext(parent, "assistant-worker", "", "")
	cfg := s.options.Config()
	for shard := 0; shard < s.storage.UserDataDBCount() && parent.Err() == nil; shard++ {
		session := s.storage.UserDataDBByIndex(shard).NewAssistantSession(ctx)
		var candidates []models.AssistantMessage
		err := session.Cols("id", "uid", "created_at", "run_state", session.Engine().GetColumnMapper().Obj2Table("RuntimeJSON"), "lease_expires_at").Where("id > ? AND role = ? AND (run_state = ? OR (run_state IN (?, ?, ?) AND lease_expires_at <= ?))", s.scanAfter[shard], "assistant", models.AssistantRunQueued, models.AssistantRunPreparing, models.AssistantRunRunning, models.AssistantRunCancelRequested, s.storage.now().UnixMilli()).Asc("id").Limit(int(cfg.ScanBatchSize)).Find(&candidates)
		session.Close()
		if err != nil {

			continue
		}
		if len(candidates) == 0 {
			s.scanAfter[shard] = 0
		}
		for _, candidate := range candidates {
			s.scanAfter[shard] = candidate.Id
			if parent.Err() != nil {
				return
			}
			select {
			case s.slots <- struct{}{}:
			default:
				if candidate.RunState != models.AssistantRunQueued || s.storage.now().UnixMilli()-candidate.CreatedAt >= int64(s.policy(&candidate).Experience.QueueTimeoutSeconds)*1000 {
					s.audit(ctx, candidate.Uid, candidate.Id)
				}
				continue
			}
			run, err := s.claim(ctx, candidate.Uid, candidate.Id)
			if err != nil || run == nil {
				<-s.slots
				continue
			}
			work.Add(1)
			go func() { defer work.Done(); defer func() { <-s.slots }(); s.execute(parent, run) }()
		}
	}
}
func (s *AssistantRunService) audit(ctx core.Context, uid, id int64) {
	_ = s.storage.WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		run, err := assistantConversationMessage(tx.Session, uid, id)
		if err != nil {
			return err
		}
		return s.expire(tx, run)
	})
}
func (s *AssistantRunService) expire(tx *AssistantStorageTx, run *models.AssistantMessage) error {
	if run.RunState.IsTerminal() {
		return nil
	}
	cfg := s.policy(run).Experience
	now := tx.Now.UnixMilli()
	if run.RunState == models.AssistantRunQueued && now-run.CreatedAt >= int64(cfg.QueueTimeoutSeconds)*1000 {
		return s.finish(tx, run, models.AssistantRunFailed, models.AssistantReasonQueueTimeout)
	}
	if run.RunState != models.AssistantRunQueued && run.LeaseExpiresAt <= now {
		return s.finish(tx, run, models.AssistantRunInterrupted, models.AssistantReasonInterrupted)
	}
	return nil
}
func (s *AssistantRunService) claim(ctx core.Context, uid, id int64) (*models.AssistantMessage, error) {
	var claimed *models.AssistantMessage
	err := s.storage.WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		claimed = nil
		run, err := assistantConversationMessage(tx.Session, uid, id)
		if err != nil {
			return err
		}
		if err = s.expire(tx, run); err != nil {
			return err
		}
		if run.RunState != models.AssistantRunQueued {
			return nil
		}
		if s.options.GenerationAvailable == nil || !s.options.GenerationAvailable(ctx, uid) {
			return s.finish(tx, run, models.AssistantRunFailed, models.AssistantReasonGenerationUnavailable)
		}
		var limits models.AssistantLimits
		if json.Unmarshal([]byte(run.LimitsJSON), &limits) != nil {
			return s.finish(tx, run, models.AssistantRunFailed, models.AssistantReasonStorageUnavailable)
		}
		if err = tx.MarkRunRunning(run, &limits); err != nil {
			return err
		}
		run.Owner = s.owner
		run.LeaseEpoch++
		run.RunState = models.AssistantRunPreparing
		run.LeaseExpiresAt = tx.Now.Add(time.Duration(s.policy(run).Experience.LeaseSeconds) * time.Second).UnixMilli()
		run.Deadline = tx.Now.Add(time.Duration(s.policy(run).RunTimeout) * time.Millisecond).UnixMilli()
		run.SnapshotRevision++
		run.SnapshotSavedAt = tx.Now.UnixMilli()
		if err = assistantRunSave(tx, run); err != nil {
			return err
		}
		claimed = run
		return nil
	})
	if err == nil && claimed != nil {
		s.observe(ctx, claimed)
	}
	return claimed, err
}
func assistantRunSave(tx *AssistantStorageTx, run *models.AssistantMessage) error {
	_, err := tx.Session.Where("uid = ? AND id = ?", tx.State.Uid, run.Id).AllCols().Update(run)
	return err
}
func (s *AssistantRunService) finish(tx *AssistantStorageTx, run *models.AssistantMessage, state models.AssistantRunState, reason models.AssistantReasonCode) error {
	if run.RunState.IsTerminal() {
		return nil
	}
	conversation, err := assistantConversationLoad(tx.Session, run.Uid, run.ConversationId)
	if err != nil {
		return err
	}
	question, err := assistantConversationMessage(tx.Session, run.Uid, run.UserMessageId)
	if err != nil {
		return err
	}
	run.RunState = state
	run.ReasonCode = reason
	run.CompletedAt = tx.Now.UnixMilli()
	run.SnapshotSavedAt = run.CompletedAt
	run.SnapshotRevision++
	n, err := AssistantAttemptStorageBytes(question, run)
	if err != nil {
		return err
	}
	if err = tx.SettleRun(conversation, run, n); err != nil {
		return err
	}
	if conversation.ActiveRunId == run.Id {
		conversation.ActiveRunId = 0
	}
	conversation.LastRunState = state
	conversation.LastReasonCode = reason
	conversation.ActivityAt = run.CompletedAt
	if state == models.AssistantRunCompleted {
		question.SelectedAssistantMessageId = run.Id
		if err = assistantRunSave(tx, question); err != nil {
			return err
		}
	}
	if err = assistantRunSave(tx, run); err != nil {
		return err
	}
	return assistantConversationSave(tx, conversation)
}
