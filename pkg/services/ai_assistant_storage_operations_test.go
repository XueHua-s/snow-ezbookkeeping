package services

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

func TestAssistantStorageExternalOperationAcceptance(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			s, _ := assistantStorageFixture(t, kind)
			other := *s
			ctx := core.NewNullContext()
			cmd := storageCommand(s, "whole-user-clear", 0)
			var owners atomic.Int64
			var group sync.WaitGroup
			start := make(chan struct{})
			for i := 0; i < 12; i++ {
				group.Add(1)
				go func(i int) {
					defer group.Done()
					<-start
					instance := s
					if i%2 == 1 {
						instance = &other
					}
					_, created, err := instance.ExecuteCommandWithAcceptance(ctx, 7, "user_data.clear_assistant", cmd, cmd, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
						tx.State.DeletionEpoch++
						return models.AssistantCommandResult{}, nil
					})
					if err != nil {
						t.Errorf("acceptance failed: %v", err)
					}
					if created {
						owners.Add(1)
					}
				}(i)
			}
			close(start)
			group.Wait()
			if owners.Load() != 1 {
				t.Fatalf("external-operation owners=%d", owners.Load())
			}
			state, err := s.GetUserState(ctx, 7)
			if err != nil || state.DeletionEpoch != 1 {
				t.Fatal("clear replayed", err)
			}
			completed, err := s.GetCommandOperationCompleted(ctx, 7, cmd.ClientRequestId)
			if err != nil || completed {
				t.Fatal("unfinished external operation reported complete", err)
			}
			result, err := s.GetCommand(ctx, 7, cmd.ClientRequestId)
			if err != nil || result.OperationCompleted == nil || *result.OperationCompleted {
				t.Fatal("confirmation lost incomplete status", err)
			}
			for i := 0; i < 2; i++ {
				if err := s.CompleteCommandOperation(ctx, 7, cmd.ClientRequestId); err != nil {
					t.Fatal(err)
				}
			}
			result, err = s.GetCommand(ctx, 7, cmd.ClientRequestId)
			if err != nil || result.OperationCompleted == nil || !*result.OperationCompleted {
				t.Fatal("completion not durable", err)
			}
			_, created, err := s.ExecuteCommandWithAcceptance(ctx, 7, "user_data.clear_assistant", cmd, cmd, func(tx *AssistantStorageTx) (models.AssistantCommandResult, error) {
				t.Fatal("completed operation replayed")
				return models.AssistantCommandResult{}, nil
			})
			if err != nil || created {
				t.Fatal("replay claimed operation ownership", err)
			}
			if _, err := s.GetCommandOperationCompleted(ctx, 8, cmd.ClientRequestId); err == nil {
				t.Fatal("foreign completion readable")
			}
		})
	}
}
