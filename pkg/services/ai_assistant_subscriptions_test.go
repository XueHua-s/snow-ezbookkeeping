package services

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
)

func TestAssistantSubscriptionDatabaseLeases(t *testing.T) {
	for _, kind := range []string{settings.Sqlite3DbType, settings.MySqlDbType, settings.PostgresDbType} {
		t.Run(kind, func(t *testing.T) {
			storage, engine := assistantStorageFixture(t, kind)
			otherInstance := *storage
			_, err := engine.Insert(&models.AssistantConversation{Id: 1, Uid: 7}, &models.AssistantMessage{Id: 2, Uid: 7, UserMessageId: 2, ConversationId: 1, Role: "assistant"}, &models.AssistantMessage{Id: 3, Uid: 7, UserMessageId: 3, ConversationId: 1, Role: "assistant"}, &models.AssistantMessage{Id: 4, Uid: 7, UserMessageId: 4, ConversationId: 1, Role: "assistant"})
			require.NoError(t, err)
			ctx := core.NewNullContext()
			var mu sync.Mutex
			var leases []string
			var group sync.WaitGroup
			for i := 0; i < 12; i++ {
				group.Add(1)
				go func(index int) {
					defer group.Done()
					instance := storage
					if index%2 == 0 {
						instance = &otherInstance
					}
					id, err := instance.AcquireSubscription(ctx, 7, "2", time.Minute)
					mu.Lock()
					defer mu.Unlock()
					if err == nil {
						leases = append(leases, id)
					} else if !errors.Is(err, models.AssistantReasonSubscriptionLimit) {
						t.Errorf("unexpected claim: %v", err)
					}
				}(i)
			}
			group.Wait()
			require.Len(t, leases, 4)
			for i := 0; i < 4; i++ {
				_, err := otherInstance.AcquireSubscription(ctx, 7, "3", time.Minute)
				require.NoError(t, err)
			}
			_, err = storage.AcquireSubscription(ctx, 7, "4", time.Minute)
			require.ErrorIs(t, err, models.AssistantReasonSubscriptionLimit)
			_, err = storage.AcquireSubscription(ctx, 8, "2", time.Minute)
			require.ErrorIs(t, err, models.AssistantReasonNotAccessible)
			require.NoError(t, otherInstance.ReleaseSubscription(ctx, 7, "2", leases[0]))
			require.NoError(t, otherInstance.ReleaseSubscription(ctx, 7, "2", leases[0]))
			_, err = storage.AcquireSubscription(ctx, 7, "4", time.Minute)
			require.NoError(t, err)
			now := storage.now()
			storage.now = func() time.Time { return now.Add(61 * time.Second) }
			otherInstance.now = storage.now
			require.ErrorIs(t, storage.RenewSubscription(ctx, 7, "2", leases[1], time.Minute), models.AssistantReasonSubscriptionLimit)
			id, err := otherInstance.AcquireSubscription(ctx, 7, "2", time.Minute)
			require.NoError(t, err)
			require.NoError(t, storage.RenewSubscription(ctx, 7, "2", id, time.Minute))
			state, err := storage.GetUserState(ctx, 7)
			require.NoError(t, err)
			require.Less(t, len(state.SubscriptionLeasesJSON), 512)
		})
	}
}
