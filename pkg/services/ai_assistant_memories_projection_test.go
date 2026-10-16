package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
)

func TestAssistantMemoryListExportContract(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) { assistantMemoryListExportContract(t, kind) })
	}
}

func assistantMemoryListExportContract(t *testing.T, kind string) {
	conv := assistantConversationFixture(t, kind)
	ctx := core.NewNullContext()
	memory := NewAssistantMemoryService(conv.storage, func() *settings.Config { return &settings.Config{} })
	require.NoError(t, conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		tx.State.ActiveMemoryCount = 3
		for i, sources := range []string{"", "null", `[{"conversationId":"10","version":1},{"conversationId":"11","messageId":"12","version":2}]`} {
			_, err := tx.Session.Insert(&models.AssistantMemory{Id: int64(i + 1), Uid: 7, Key: fmt.Sprint(i), Text: "preference", Status: "active", SourcesJSON: sources, SchemaVersion: 0})
			if err != nil {
				return err
			}
		}
		return nil
	}))
	list, err := memory.List(ctx, 7)
	require.NoError(t, err)
	var out bytes.Buffer
	require.NoError(t, conv.Export(ctx, 7, models.AssistantExportRequest{All: true, Format: "json"}, &out))
	var exported struct {
		Memories []models.AssistantMemoryDTO `json:"memories"`
	}
	require.NoError(t, json.Unmarshal(out.Bytes(), &exported))
	require.Equal(t, list.Items, exported.Memories)
	require.Len(t, list.Items, 3)
	for _, item := range list.Items {
		require.Equal(t, models.AssistantSchemaVersion, item.SchemaVersion)
		require.NotNil(t, item.Sources)
		require.NotNil(t, item.SourceMessageIds)
	}
	require.Equal(t, []string{"12"}, list.Items[2].SourceMessageIds)
	require.Len(t, list.Items[2].Sources, 2)
	// A corrupt row must not block the manager, but exporting it must fail visibly.
	require.NoError(t, conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		var row models.AssistantMemory
		_, err := tx.Session.Where("uid = ? AND id = ?", 7, 1).Get(&row)
		if err != nil {
			return err
		}
		row.SourcesJSON = "{"
		_, err = tx.Session.Where("uid = ? AND id = ?", 7, 1).AllCols().Update(&row)
		return err
	}))
	list, err = memory.List(ctx, 7)
	require.NoError(t, err)
	require.Len(t, list.Items, 2)
	out.Reset()
	require.ErrorIs(t, conv.Export(ctx, 7, models.AssistantExportRequest{All: true, Format: "json"}, &out), models.AssistantReasonStorageUnavailable)
}

func TestAssistantMemoryRevocationBeyondDefaultConcurrency(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			for _, action := range []string{"disable", "invalidate-source"} {
				t.Run(action, func(t *testing.T) { assistantMemoryRevocationBeyondDefaultConcurrency(t, kind, action) })
			}
		})
	}
}

func assistantMemoryRevocationBeyondDefaultConcurrency(t *testing.T, kind, action string) {
	conv := assistantConversationFixture(t, kind)
	cfg := settings.DefaultAIAssistantExperienceConfig()
	cfg.Limits.QueuedPerUser = 6
	conv.options.Config = func() settings.AIAssistantExperienceConfig { return cfg }
	ctx := core.NewNullContext()
	memory := NewAssistantMemoryService(conv.storage, func() *settings.Config { return &settings.Config{} })
	_, err := memory.Settings(ctx, 7, models.AssistantMemorySettingsRequest{AssistantCommand: storageCommand(conv.storage, "enable", 0), ExpectedSettingsVersion: 1, Enabled: true})
	require.NoError(t, err)
	var runs []int64
	for i := 0; i < 6; i++ {
		sent := conversationSend(t, conv, fmt.Sprint("run", i), "", 0, "question")
		id, err := assistantParseID(sent.RunId)
		require.NoError(t, err)
		runs = append(runs, id)
	}
	require.NoError(t, conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		raw, err := json.Marshal(models.AssistantInputDependencies{MemoryMode: "on", Memories: []models.AssistantMemoryDependency{{Id: "99", Version: 1}}})
		if err != nil {
			return err
		}
		for _, id := range runs {
			run, err := assistantConversationMessage(tx.Session, 7, id)
			if err != nil {
				return err
			}
			run.InputDependenciesJSON = string(raw)
			if _, err = tx.Session.Where("uid = ? AND id = ?", 7, id).AllCols().Update(run); err != nil {
				return err
			}
		}
		return nil
	}))
	if action == "disable" {
		_, err = memory.Settings(ctx, 7, models.AssistantMemorySettingsRequest{AssistantCommand: storageCommand(conv.storage, "disable", 0), ExpectedSettingsVersion: 2, Enabled: false})
	} else {
		err = conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error { return tx.InvalidateSources(999) })
	}

	require.NoError(t, err)
	require.NoError(t, conv.storage.WithUserTransaction(ctx, 7, func(tx *AssistantStorageTx) error {
		for _, id := range runs {
			run, err := assistantConversationMessage(tx.Session, 7, id)
			if err != nil {
				return err
			}
			require.Equal(t, models.AssistantRunCanceled, run.RunState)
			require.True(t, run.ReservationSettled)
		}
		require.Zero(t, tx.State.QueuedCount)
		return nil
	}))
}
