package services

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"xorm.io/xorm"
)

// ExecuteCommand commits its receipt and business mutation together. A receipt
// lookup precedes time/epoch checks so old accepted commands remain confirmable.
func (s *AssistantStorageService) ExecuteCommand(ctx core.Context, uid int64, commandType string, command models.AssistantCommand, request any, apply func(*AssistantStorageTx) (models.AssistantCommandResult, error)) (models.AssistantCommandResult, error) {
	result, _, err := s.executeCommand(ctx, uid, commandType, command, request, apply, false)
	return result, err
}

// ExecuteCommandWithAcceptance reports ownership of the newly committed receipt.
// External operations run only when created is true; an uncertain operation is
// never automatically replayed by a second caller.
func (s *AssistantStorageService) ExecuteCommandWithAcceptance(ctx core.Context, uid int64, commandType string, command models.AssistantCommand, request any, apply func(*AssistantStorageTx) (models.AssistantCommandResult, error)) (models.AssistantCommandResult, bool, error) {
	return s.executeCommand(ctx, uid, commandType, command, request, apply, true)
}

func (s *AssistantStorageService) executeCommand(ctx core.Context, uid int64, commandType string, command models.AssistantCommand, request any, apply func(*AssistantStorageTx) (models.AssistantCommandResult, error), operationPending bool) (models.AssistantCommandResult, bool, error) {
	var result models.AssistantCommandResult
	created := false
	if commandType == "" || len(commandType) > 64 || strings.TrimSpace(command.ClientRequestId) == "" || len(command.ClientRequestId) > 128 {
		return result, false, models.AssistantReasonConversationConflict
	}
	hash, err := models.AssistantCanonicalRequestHash(commandType, struct {
		Command models.AssistantCommand
		Request any
	}{command, request})
	if err != nil {
		return result, false, models.AssistantReasonConversationConflict
	}
	err = s.WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		created = false
		if tx.State.AccountUnavailable {
			return models.AssistantReasonNotAccessible
		}
		receipt := &models.AssistantCommandReceipt{}
		found, err := tx.Session.Where("uid = ? AND client_request_id = ?", uid, command.ClientRequestId).Get(receipt)
		if err != nil {
			return err
		}
		if found {
			if receipt.CanonicalRequestHash != hash || receipt.CommandType != commandType {
				return models.AssistantReasonConversationConflict
			}
			result, err = assistantReceiptResult(tx.Session, receipt)
			return err
		}
		if err := command.ValidateTime(tx.Now, time.Duration(s.config().CommandMaxAgeHours)*time.Hour, time.Duration(s.config().CommandFutureSkewSeconds)*time.Second); err != nil {
			return models.AssistantReasonConversationConflict
		}
		if command.ExpectedDeletionEpoch != tx.State.DeletionEpoch {
			return models.AssistantReasonDataEpochConflict
		}
		result, err = apply(tx)
		if err != nil {
			return err
		}
		result.ClientRequestId, result.CommandType = command.ClientRequestId, commandType
		if result.Status == "" {
			result.Status = "accepted"
		}
		result.DeletionEpoch = tx.State.DeletionEpoch
		if commandType == "user_data.clear_assistant" {
			completed := !operationPending
			result.OperationCompleted = &completed
		}
		raw, err := json.Marshal(result)
		if err != nil {
			return err
		}
		conversationId, _ := strconv.ParseInt(result.ConversationId, 10, 64)
		memoryId, _ := strconv.ParseInt(result.MemoryId, 10, 64)
		receipt = &models.AssistantCommandReceipt{Uid: uid, ClientRequestId: command.ClientRequestId, CommandType: commandType, CanonicalRequestHash: hash, RequestIssuedAt: command.RequestIssuedAt, ExpectedDeletionEpoch: command.ExpectedDeletionEpoch, ResultJSON: string(raw), ConversationId: conversationId, MemoryId: memoryId, CreatedAt: tx.Now.UnixMilli(), ExpiresAt: tx.Now.Add(tx.ReceiptRetention).UnixMilli(), OperationCompleted: !operationPending}
		_, err = tx.Session.Insert(receipt)
		created = err == nil
		return err
	})
	return result, created && err == nil, err
}

func (s *AssistantStorageService) GetCommand(ctx core.Context, uid int64, requestId string) (models.AssistantCommandResult, error) {
	result := models.AssistantCommandResult{ClientRequestId: requestId, Status: "unknown"}
	err := s.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		receipt := &models.AssistantCommandReceipt{}
		found, err := session.Where("uid = ? AND client_request_id = ?", uid, requestId).Get(receipt)
		if err != nil || !found {
			return err
		}
		result, err = assistantReceiptResult(session, receipt)
		return err
	})
	return result, err
}

func assistantReceiptResult(session *xorm.Session, receipt *models.AssistantCommandReceipt) (models.AssistantCommandResult, error) {
	var result models.AssistantCommandResult
	if err := json.Unmarshal([]byte(receipt.ResultJSON), &result); err != nil {
		return result, models.AssistantReasonStorageUnavailable
	}
	if receipt.CommandType == "user_data.clear_assistant" {
		completed := receipt.OperationCompleted
		result.OperationCompleted = &completed
	}
	deleted := receipt.Tombstone
	if receipt.ConversationId > 0 && !deleted {
		found, err := session.Where("uid = ? AND id = ? AND deleted = ?", receipt.Uid, receipt.ConversationId, false).Exist(&models.AssistantConversation{})
		if err != nil {
			return result, err
		}
		deleted = !found
	}
	if receipt.MemoryId > 0 && !deleted {
		found, err := session.Where("uid = ? AND id = ? AND excluded = ?", receipt.Uid, receipt.MemoryId, false).Exist(&models.AssistantMemory{})
		if err != nil {
			return result, err
		}
		deleted = !found
	}
	if deleted {
		result.Status = "deleted"
	}
	return result, nil
}

// TombstoneReceipts extends retention from deletion, not original creation.
// Call in the same transaction before removing an object or its body.
func (tx *AssistantStorageTx) TombstoneReceipts(conversationId, memoryId int64, all bool) error {
	query := tx.Session.Where("uid = ?", tx.State.Uid)
	if !all {
		if conversationId > 0 {
			query = query.And("conversation_id = ?", conversationId)
		} else if memoryId > 0 {
			query = query.And("memory_id = ?", memoryId)
		} else {
			return models.AssistantReasonNotAccessible
		}
	}
	_, err := query.Cols("tombstone", "expires_at").Update(&models.AssistantCommandReceipt{Tombstone: true, ExpiresAt: tx.Now.Add(tx.ReceiptRetention).UnixMilli()})
	return err
}

// PruneReceipts excludes live-object receipts in SQL so they cannot starve the
// bounded batch of expired tombstones and objectless administrative commands.
func (s *AssistantStorageService) PruneReceipts(ctx core.Context, uid int64, limit int) error {
	if limit < 1 || limit > 100 {
		limit = 100
	}
	return s.WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		var receipts []models.AssistantCommandReceipt
		if err := tx.Session.Where("uid = ? AND expires_at < ? AND (tombstone = ? OR (conversation_id = 0 AND memory_id = 0))", uid, tx.Now.UnixMilli(), true).Asc("expires_at", "client_request_id").Limit(limit).Find(&receipts); err != nil {
			return err
		}
		for _, receipt := range receipts {
			if _, err := tx.Session.Where("uid = ? AND client_request_id = ?", uid, receipt.ClientRequestId).Delete(&models.AssistantCommandReceipt{}); err != nil {
				return err
			}
		}
		return nil
	})
}
