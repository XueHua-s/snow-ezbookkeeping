package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

// CompleteCommandOperation records completion of an already accepted external
// operation. It does not execute or retry that operation.
func (s *AssistantStorageService) CompleteCommandOperation(ctx core.Context, uid int64, requestId string) error {
	return s.WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		receipt := &models.AssistantCommandReceipt{}
		found, err := tx.Session.Where("uid = ? AND client_request_id = ?", uid, requestId).Get(receipt)
		if err != nil {
			return err
		}
		if !found {
			return models.AssistantReasonNotAccessible
		}
		if receipt.OperationCompleted {
			return nil
		}
		_, err = tx.Session.Where("uid = ? AND client_request_id = ?", uid, requestId).Cols("operation_completed").Update(&models.AssistantCommandReceipt{OperationCompleted: true})
		return err
	})
}

func (s *AssistantStorageService) GetCommandOperationCompleted(ctx core.Context, uid int64, requestId string) (bool, error) {
	completed := false
	err := s.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		receipt := &models.AssistantCommandReceipt{}
		found, err := session.Where("uid = ? AND client_request_id = ?", uid, requestId).Get(receipt)
		if err != nil {
			return err
		}
		if !found {
			return models.AssistantReasonNotAccessible
		}
		completed = receipt.OperationCompleted
		return nil
	})
	return completed, err
}
