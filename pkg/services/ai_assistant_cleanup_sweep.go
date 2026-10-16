package services

import (
	"sync"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// AssistantCleanupService retains only per-shard scan positions. Sweep is called
// independently of model availability and never loads a full user's history.
type AssistantCleanupService struct {
	storage *AssistantStorageService
	mu      sync.Mutex
	after   map[int]int64
}

func NewAssistantCleanupService(storage *AssistantStorageService) *AssistantCleanupService {
	return &AssistantCleanupService{storage: storage, after: make(map[int]int64)}
}
func (s *AssistantCleanupService) Sweep(ctx core.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.storage.now()
	candidateCutoff := now.Add(-time.Duration(s.storage.config().MemoryCandidateDays) * 24 * time.Hour).UnixMilli()
	// Separate keysets prevent a large tombstone backlog from starving receipt
	// expiry or candidates. Empty user states are never visited or rewritten.
	for shard := 0; shard < s.storage.UserDataDBCount(); shard++ {
		seen := make(map[int64]bool)
		for kind := 0; kind < 3; kind++ {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			cursor := shard*3 + kind
			session := s.storage.UserDataDBByIndex(shard).NewAssistantSession(ctx)
			switch kind {
			case 0:
				session.Table(&models.AssistantConversation{}).Where("deleted = ?", true)
			case 1:
				session.Table(&models.AssistantMemory{}).Where("status = ? AND (created_at <= ? OR (expires_at > 0 AND expires_at <= ?))", "candidate", candidateCutoff, now.UnixMilli())
			case 2:
				session.Table(&models.AssistantCommandReceipt{}).Where("expires_at < ? AND (tombstone = ? OR (conversation_id = 0 AND memory_id = 0))", now.UnixMilli(), true)
			}
			var users []struct{ Uid int64 }
			err := session.Distinct("uid").And("uid > ?", s.after[cursor]).Asc("uid").Limit(50).Find(&users)
			session.Close()
			if err != nil {
				return assistantStorageSafeError(err)
			}
			if len(users) == 0 {
				s.after[cursor] = 0
				continue
			}
			for _, user := range users {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				if !seen[user.Uid] {
					if err = s.CleanupUser(ctx, user.Uid); err != nil {
						return err
					}
					seen[user.Uid] = true
				}
				s.after[cursor] = user.Uid
			}
			log.Infof(ctx, "[assistant] cleanup batch shard=%d kind=%d users=%d batch_full=%t", shard, kind, len(users), len(users) == 50)
		}
	}
	return nil
}

// CleanupUser performs bounded physical erasure after logical deletion. An
// unfinished conversation remains a tombstone until its last message is gone.
func (s *AssistantCleanupService) CleanupUser(ctx core.Context, uid int64) error {
	err := s.storage.WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		var conversations []models.AssistantConversation
		if err := tx.Session.Where("uid = ? AND deleted = ?", uid, true).Asc("id").Limit(20).Find(&conversations); err != nil {
			return err
		}
		remaining := 1000
		for _, conversation := range conversations {
			// This is operational metadata only: never include a title or message body.
			if conversation.DeletedAt > 0 && conversation.DeletedAt < tx.Now.Add(-time.Duration(s.storage.config().CleanupHours)*time.Hour).UnixMilli() {
				log.Warnf(ctx, "[assistant] cleanup overdue uid=%d conversation=%d", uid, conversation.Id)
			}
			if remaining == 0 {
				break
			}
			var messages []models.AssistantMessage
			if err := tx.Session.Cols("id").Where("uid = ? AND conversation_id = ?", uid, conversation.Id).Asc("id").Limit(remaining).Find(&messages); err != nil {
				return err
			}
			if len(messages) > 0 {
				ids := make([]int64, len(messages))
				for i, m := range messages {
					ids[i] = m.Id
				}
				if _, err := tx.Session.Where("uid = ?", uid).In("id", ids).Delete(&models.AssistantMessage{}); err != nil {
					return err
				}
				remaining -= len(messages)
			}
			count, err := tx.Session.Where("uid = ? AND conversation_id = ?", uid, conversation.Id).Count(&models.AssistantMessage{})
			if err != nil {
				return err
			}
			if count == 0 {
				if _, err = tx.Session.Where("uid = ? AND id = ? AND deleted = ?", uid, conversation.Id, true).Delete(&models.AssistantConversation{}); err != nil {
					return err
				}
			}
		}
		var memories []models.AssistantMemory
		cutoff := tx.Now.Add(-time.Duration(s.storage.config().MemoryCandidateDays) * 24 * time.Hour).UnixMilli()
		if err := tx.Session.Where("uid = ? AND status = ? AND (created_at <= ? OR (expires_at > 0 AND expires_at <= ?))", uid, "candidate", cutoff, tx.Now.UnixMilli()).Asc("id").Limit(100).Find(&memories); err != nil {
			return err
		}
		for _, memory := range memories {
			if err := tx.TombstoneReceipts(0, memory.Id, false); err != nil {
				return err
			}
			if _, err := tx.Session.Where("uid = ? AND id = ?", uid, memory.Id).Delete(&models.AssistantMemory{}); err != nil {
				return err
			}
			tx.State.UsedBytes -= memory.UsedBytes
			tx.State.CandidateMemoryCount--
		}
		return nil
	})
	if err != nil {
		return err
	}
	return s.storage.PruneReceipts(ctx, uid, 100)
}
