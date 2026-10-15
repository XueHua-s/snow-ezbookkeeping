package api

import (
	"context"
	"sync"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

type assistantRequestClass uint8

const (
	assistantGenerationRequest assistantRequestClass = iota
	assistantSearchExportRequest
	assistantStatusRequest
)

type assistantRateWindow struct {
	minute int64
	counts [3]int
}
type assistantRequestLimiter struct {
	mu    sync.Mutex
	users map[int64]assistantRateWindow
	heavy int
}

// Limits are per process. Durable generation admission and quotas remain in the
// database, so a second server cannot bypass the authoritative queue boundary.
// Only state polling is throttled; cancel/delete and reading history remain free.
func (l *assistantRequestLimiter) acquire(uid int64, class assistantRequestClass, limits models.AssistantLimits, now time.Time) (func(), bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	minute := now.Unix() / 60
	window, found := l.users[uid]
	if !found {
		if l.users == nil {
			l.users = make(map[int64]assistantRateWindow)
		}
		if len(l.users) >= limits.RateLimitUsers {
			for key, old := range l.users {
				if old.minute < minute {
					delete(l.users, key)
				}
			}
		}
		if len(l.users) >= limits.RateLimitUsers {
			return nil, false
		}
	}
	if window.minute != minute {
		window = assistantRateWindow{minute: minute}
	}
	allowance := []int{limits.GenerationRequestsPerMinute, limits.SearchExportRequestsPerMinute, limits.StatusRequestsPerMinute}[class]
	if window.counts[class] >= allowance || class == assistantSearchExportRequest && l.heavy >= limits.SearchExportConcurrency {
		return nil, false
	}
	window.counts[class]++
	l.users[uid] = window
	if class == assistantSearchExportRequest {
		l.heavy++
	}
	var once sync.Once
	return func() {
		once.Do(func() {
			if class == assistantSearchExportRequest {
				l.mu.Lock()
				l.heavy--
				l.mu.Unlock()
			}
		})
	}, true
}
func (a *LargeLanguageModelsApi) assistantLimit(c *core.WebContext, class assistantRequestClass, export bool) (func(), *errs.Error) {
	limits := a.CurrentConfig().GetAIAssistantExperienceConfig().Limits
	release, ok := a.experience.limiter.acquire(c.GetCurrentUid(), class, limits, time.Now())
	if !ok {
		log.Infof(c, "[assistant] request_rejected uid=%d class=%d reason=queue_full", c.GetCurrentUid(), class)
		c.Header("Retry-After", "60")
		return nil, assistantExperienceError(models.AssistantReasonQueueFull)
	}
	if class == assistantSearchExportRequest {
		timeout := limits.SearchTimeoutSeconds
		if export {
			timeout = limits.ExportTimeoutSeconds
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Duration(timeout)*time.Second)
		c.Request = c.Request.WithContext(ctx)
		return func() { cancel(); release() }, nil
	}
	return release, nil
}
