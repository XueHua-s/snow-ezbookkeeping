package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

type assistantSubscriptionLease struct {
	ID        string `json:"id"`
	RunID     string `json:"runId"`
	ExpiresAt int64  `json:"expiresAt"`
}

// Subscription slots are bounded metadata under the same database CAS as the
// user's data. They coordinate API instances without coupling sockets to workers.
func (s *AssistantStorageService) AcquireSubscription(ctx core.Context, uid int64, runID string, duration time.Duration) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", models.AssistantReasonStorageUnavailable
	}
	id := hex.EncodeToString(random)
	err := s.updateSubscription(ctx, uid, runID, id, duration, "acquire")
	return id, err
}
func (s *AssistantStorageService) RenewSubscription(ctx core.Context, uid int64, runID, id string, duration time.Duration) error {
	return s.updateSubscription(ctx, uid, runID, id, duration, "renew")
}
func (s *AssistantStorageService) ReleaseSubscription(ctx core.Context, uid int64, runID, id string) error {
	return s.updateSubscription(ctx, uid, runID, id, 0, "release")
}
func (s *AssistantStorageService) updateSubscription(ctx core.Context, uid int64, runID, id string, duration time.Duration, action string) error {
	return s.WithUserTransaction(ctx, uid, func(tx *AssistantStorageTx) error {
		var leases []assistantSubscriptionLease
		if len(tx.State.SubscriptionLeasesJSON) > 1024*1024 {
			return models.AssistantReasonStorageUnavailable
		}
		if tx.State.SubscriptionLeasesJSON != "" {
			if err := json.Unmarshal([]byte(tx.State.SubscriptionLeasesJSON), &leases); err != nil {
				return models.AssistantReasonStorageUnavailable
			}
		}
		live := leases[:0]
		found, runCount := false, 0
		for _, lease := range leases {
			if lease.ExpiresAt <= tx.Now.UnixMilli() {
				continue
			}
			if lease.ID == id && lease.RunID == runID {
				found = true
				if action == "release" {
					continue
				}
				lease.ExpiresAt = tx.Now.Add(duration).UnixMilli()
			}
			live = append(live, lease)
			if lease.RunID == runID {
				runCount++
			}
		}
		if action != "release" {
			if tx.State.AccountUnavailable {
				return models.AssistantReasonNotAccessible
			}
			run, err := strconv.ParseInt(runID, 10, 64)
			if err != nil || run <= 0 {
				return models.AssistantReasonNotAccessible
			}
			message, err := assistantConversationMessage(tx.Session, uid, run)
			if err != nil {
				return err
			}
			if message.Role != "assistant" {
				return models.AssistantReasonNotAccessible
			}
			if duration <= 0 {
				return models.AssistantReasonStorageUnavailable
			}
		}
		if action == "acquire" {
			if len(live) >= tx.Limits.SubscriptionsPerUser || runCount >= tx.Limits.SubscriptionsPerRun {
				return models.AssistantReasonSubscriptionLimit
			}
			live = append(live, assistantSubscriptionLease{ID: id, RunID: runID, ExpiresAt: tx.Now.Add(duration).UnixMilli()})
		} else if action == "renew" && !found {
			return models.AssistantReasonSubscriptionLimit
		}
		raw, err := json.Marshal(live)
		if err != nil {
			return err
		}
		tx.State.SubscriptionLeasesJSON = string(raw)
		return nil
	})
}
