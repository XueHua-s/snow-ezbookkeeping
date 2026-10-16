package services

import (
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"strconv"
	"xorm.io/xorm"
)

func (s *AssistantRunService) GetRunSnapshot(ctx core.Context, uid int64, runID string) (models.AssistantRunSnapshot, error) {
	var snapshot models.AssistantRunSnapshot
	id, err := strconv.ParseInt(runID, 10, 64)
	if err != nil || id <= 0 {
		return snapshot, models.AssistantReasonNotAccessible
	}
	err = s.storage.WithUserRead(ctx, uid, func(session *xorm.Session) error {
		run, err := assistantConversationMessage(session, uid, id)
		if err != nil {
			return err
		}
		if run.Role != "assistant" {
			return models.AssistantReasonNotAccessible
		}
		snapshot, err = AssistantMessageSnapshot(run)
		return err
	})
	return snapshot, err
}
func (s *AssistantRunService) Cancel(ctx core.Context, uid int64, request models.AssistantCancelRequest) (models.AssistantCommandResult, error) {
	return s.storage.ExecuteCommand(ctx, uid, "runs.cancel", request.AssistantCommand, request, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
		id, err := strconv.ParseInt(request.RunId, 10, 64)
		if err != nil || id <= 0 {
			return models.AssistantCommandResult{}, models.AssistantReasonNotAccessible
		}
		run, err := assistantConversationMessage(tx.Session, uid, id)
		if err != nil {
			return models.AssistantCommandResult{}, err
		}
		if run.Role != "assistant" {
			return models.AssistantCommandResult{}, models.AssistantReasonNotAccessible
		}
		result := models.AssistantCommandResult{ConversationId: assistantConversationID(run.ConversationId), RunId: request.RunId}
		if run.RunState.IsTerminal() {
			return result, nil
		}
		if run.RunState == models.AssistantRunQueued {
			return result, s.finish(tx, run, models.AssistantRunCanceled, "")
		}
		run.RunState = models.AssistantRunCancelRequested
		run.CancelRequested = true
		run.SnapshotRevision++
		run.SnapshotSavedAt = tx.Now.UnixMilli()
		return result, assistantRunSave(tx, run)
	})
}
