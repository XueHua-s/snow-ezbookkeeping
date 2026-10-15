package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func assistantSnapshotFrame(snapshot models.AssistantRunSnapshot, limit int) ([]byte, error) {
	event := "message.snapshot"
	if snapshot.State.IsTerminal() {
		event = "run.terminal"
	}
	payload, err := json.Marshal(snapshot)
	if err != nil {
		return nil, err
	}
	frame := []byte(fmt.Sprintf("event: %s\nid: %s:%d\ndata: %s\n\n", event, snapshot.RunId, snapshot.Revision, payload))
	if len(frame) > limit {
		return nil, models.AssistantReasonOutputLimit
	}
	return frame, nil
}

func (a *LargeLanguageModelsApi) AssistantRunStreamHandler(c *core.WebContext) *errs.Error {
	if err := a.assistantExperienceReady(c); err != nil {
		return err
	}
	if err := assistantQuery(c, "runId", "afterRevision"); err != nil {
		return err
	}
	runID, err := assistantQueryInt(c, "runId", 1<<63-1)
	if err != nil || runID <= 0 {
		return errs.ErrIncompleteOrIncorrectSubmission
	}
	if _, err := assistantQueryInt(c, "afterRevision", 1<<53-1); err != nil {
		return err
	}
	claims := c.GetTokenClaims()
	if claims == nil || claims.ExpiresAt <= time.Now().Unix() {
		return errs.ErrUnauthorizedAccess
	}
	config := a.CurrentConfig().GetAIAssistantExperienceConfig()
	lookup := func() (models.AssistantRunSnapshot, error) {
		return a.experience.runs.GetRunSnapshot(assistantHTTPContext(c), c.GetCurrentUid(), c.Query("runId"))
	}
	snapshot, readErr := lookup()
	if readErr != nil {
		return assistantExperienceError(readErr)
	}
	frame, frameErr := assistantSnapshotFrame(snapshot, config.Limits.FrameBytes)
	if frameErr != nil {
		return assistantExperienceError(frameErr)
	}
	leaseDuration := (3*time.Duration(config.StreamHeartbeatSeconds) + time.Duration(config.SubscriptionWriteTimeoutSeconds)) * time.Second
	leaseID, leaseErr := a.experience.storage.AcquireSubscription(assistantHTTPContext(c), c.GetCurrentUid(), snapshot.RunId, leaseDuration)
	if leaseErr != nil {
		c.Header("Retry-After", "15")
		return assistantExperienceError(leaseErr)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = a.experience.storage.ReleaseSubscription(core.NewAssistantContext(ctx, c.GetContextId(), c.GetClientLocale(), c.ClientIP()), c.GetCurrentUid(), snapshot.RunId, leaseID)
	}()
	renew := func() error {
		return a.experience.storage.RenewSubscription(assistantHTTPContext(c), c.GetCurrentUid(), snapshot.RunId, leaseID, leaseDuration)
	}
	// Reconnection always starts with the current durable projection, even when
	// afterRevision matches. It never schedules or takes ownership of generation.
	c.Header("Content-Type", "text/event-stream; charset=utf-8")
	c.Header("Cache-Control", "no-store")
	c.Header("X-Accel-Buffering", "no")
	c.Header("X-Content-Type-Options", "nosniff")
	streamAssistantSnapshots(c, snapshot, frame, lookup, config.Limits.FrameBytes,
		time.Duration(config.CheckpointMilliseconds)*time.Millisecond,
		time.Duration(config.StreamHeartbeatSeconds)*time.Second,
		time.Duration(config.SubscriptionWriteTimeoutSeconds)*time.Second,
		time.Unix(claims.ExpiresAt, 0), renew)
	return nil
}

// There is no event queue and no worker callback: a blocked subscriber holds at
// most one serialized snapshot and cannot block checkpoint persistence.
func streamAssistantSnapshots(c *core.WebContext, snapshot models.AssistantRunSnapshot, frame []byte,
	lookup func() (models.AssistantRunSnapshot, error), frameLimit int,
	pollInterval, heartbeatInterval, writeTimeout time.Duration, expiresAt time.Time, renew func() error) {
	controller := http.NewResponseController(c.Writer)
	write := func(data []byte) bool {
		if c.Request.Context().Err() != nil || !time.Now().Before(expiresAt) {
			return false
		}
		// Gin exposes Unwrap, so deadlines reach the real HTTP connection. If a
		// custom writer cannot enforce a deadline, do not open an unbounded stream.
		if controller.SetWriteDeadline(time.Now().Add(writeTimeout)) != nil {
			return false
		}
		if _, err := c.Writer.Write(data); err != nil {
			return false
		}
		return controller.Flush() == nil
	}
	defer controller.SetWriteDeadline(time.Time{})
	if !write(frame) || snapshot.State.IsTerminal() {
		return
	}
	poll := time.NewTicker(pollInterval)
	heartbeat := time.NewTicker(heartbeatInterval)
	expiration := time.NewTimer(time.Until(expiresAt))
	defer poll.Stop()
	defer heartbeat.Stop()
	defer expiration.Stop()
	revision := snapshot.Revision
	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-expiration.C:
			return
		case <-heartbeat.C:
			if renew != nil && renew() != nil {
				return
			}
			if !write([]byte(": heartbeat\n\n")) {
				return
			}
		case <-poll.C:
			next, err := lookup()
			if err != nil {
				return
			} // deletion/authorization/storage failure: synchronize via status; never fabricate done.
			if next.Revision <= revision {
				continue
			}
			frame, err := assistantSnapshotFrame(next, frameLimit)
			if err != nil || !write(frame) {
				return
			}
			revision = next.Revision
			if next.State.IsTerminal() {
				return
			}
		}
	}
}
