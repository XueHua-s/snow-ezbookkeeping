package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

// SetAssistantStorage shares the web runtime's transaction service. CLI commands
// have no runtime and use the same configured datastore without starting workers.
func (s *UserService) SetAssistantStorage(storage *AssistantStorageService) {
	s.assistantStorage = storage
}
func (s *UserService) assistantLifecycleStorage() *AssistantStorageService {
	if s.assistantStorage != nil {
		return s.assistantStorage
	}
	return NewAssistantStorageService(s.ServiceUsingDB.container)
}

func (s *UserService) fenceAssistantAccount(ctx core.Context, uid int64, unavailable, remove bool) error {
	return s.assistantLifecycleStorage().WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		tx.State.AccountUnavailable = unavailable
		if remove {
			return tx.ClearAssistantData()
		}
		if !unavailable {
			return nil
		}
		tx.State.DeletionEpoch++
		tx.State.MemoryGenerationEpoch++
		var after int64
		for {
			var rows []models.AssistantConversation
			if err := tx.Session.Where("uid = ? AND deleted = ? AND id > ?", uid, false, after).Asc("id").Limit(50).Find(&rows); err != nil {
				return err
			}
			if len(rows) == 0 {
				return nil
			}
			for _, row := range rows {
				after = row.Id
				if err := tx.CancelConversationRuns(row.Id, models.AssistantReasonProviderUnavailable); err != nil {
					return err
				}
			}
		}
	})
}
