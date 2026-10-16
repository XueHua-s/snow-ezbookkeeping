package api

import (
	"context"
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/stretchr/testify/require"
)

func TestClearAllDataAssistantLegacyReplayFence(t *testing.T) {
	_, engine := harnessAPIFixture(t)
	require.NoError(t, engine.Sync2(&models.AssistantConversation{}, &models.AssistantMessage{}, &models.AssistantMemory{}, &models.AssistantUserState{}, &models.AssistantCommandReceipt{}))
	storage := services.NewAssistantStorageService(datastore.Container)
	a := &DataManagementsApi{assistantStorage: storage}
	c := harnessContext("", 7, context.Background(), httptest.NewRecorder())
	claims := &core.UserTokenClaims{Uid: 7, UserTokenId: "private-token-canary", IssuedAt: time.Now().Unix()}
	c.SetTokenClaims(claims)
	clear := func(command *models.AssistantCommand) error {
		_, _, err := a.beginAssistantForAllData(c, command)
		return err
	}
	require.NoError(t, clear(nil))
	state, err := storage.GetUserState(c, 7)
	require.NoError(t, err)
	require.EqualValues(t, 1, state.DeletionEpoch)
	require.NoError(t, storage.WithUserTransaction(c, 7, func(tx *services.AssistantStorageTx) error {
		tx.State.ConversationCount = 1
		_, err := tx.Session.Insert(&models.AssistantConversation{Id: 777, Uid: 7, Title: "new history survives"})
		return err
	}))
	require.NoError(t, clear(nil))
	count, err := engine.Where("uid = ? AND deleted = ?", 7, false).Count(&models.AssistantConversation{})
	require.NoError(t, err)
	require.EqualValues(t, 1, count)
	var receipts []models.AssistantCommandReceipt
	require.NoError(t, engine.Find(&receipts))
	require.Len(t, receipts, 1)
	require.NotContains(t, receipts[0].ClientRequestId, "private-token-canary")
	require.NotContains(t, receipts[0].ResultJSON, "new history survives")
	claims.UserTokenId = "different-legacy-token"
	require.ErrorIs(t, clear(nil), models.AssistantReasonDataEpochConflict)
	explicit := &models.AssistantCommand{ClientRequestId: "explicit-clear", RequestIssuedAt: time.Now().UTC().Format(time.RFC3339Nano), ExpectedDeletionEpoch: 1}
	require.NoError(t, clear(explicit))
	state, err = storage.GetUserState(c, 7)
	require.NoError(t, err)
	require.EqualValues(t, 2, state.DeletionEpoch)
}

func TestClearAllDataReplayNeverReentersLedgerServices(t *testing.T) {
	_, engine := harnessAPIFixture(t)
	require.NoError(t, engine.Sync2(&models.AssistantConversation{}, &models.AssistantMessage{}, &models.AssistantMemory{}, &models.AssistantUserState{}, &models.AssistantCommandReceipt{}))
	storage := services.NewAssistantStorageService(datastore.Container)
	// Nil ledger services are intentional: any replay that reaches them panics.
	a := &DataManagementsApi{assistantStorage: storage, users: services.Users}
	_, err := engine.Where("uid = ?", 7).Cols("password", "salt").Update(&models.User{Password: utils.EncodePassword("test-password", "salt"), Salt: "salt"})
	require.NoError(t, err)
	command := models.AssistantCommand{ClientRequestId: "global-clear", RequestIssuedAt: time.Now().UTC().Format(time.RFC3339Nano)}
	contextForRequest := func() *core.WebContext {
		raw, err := json.Marshal(models.ClearDataRequest{Password: "test-password", AssistantCommand: &command})
		require.NoError(t, err)
		c := harnessContext(string(raw), 7, context.Background(), httptest.NewRecorder())
		c.SetTokenClaims(&core.UserTokenClaims{Uid: 7, Type: core.USER_TOKEN_TYPE_NORMAL})
		return c
	}
	c := contextForRequest()
	_, created, err := a.beginAssistantForAllData(c, &command)
	require.NoError(t, err)
	require.True(t, created)
	_, apiErr := a.ClearAllDataHandler(contextForRequest())
	require.Same(t, errs.ErrUserDataClearIncomplete, apiErr)
	require.NoError(t, storage.CompleteCommandOperation(c, 7, command.ClientRequestId))
	result, apiErr := a.ClearAllDataHandler(contextForRequest())
	require.Nil(t, apiErr)
	require.Equal(t, true, result)
}
