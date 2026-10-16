package api

import (
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestAssistantExperienceRequestLimits(t *testing.T) {
	limits := settings.DefaultAIAssistantExperienceConfig().Limits
	limits.GenerationRequestsPerMinute = 2
	limits.StatusRequestsPerMinute = 3
	limits.SearchExportConcurrency = 1
	limits.RateLimitUsers = 2
	var limiter assistantRequestLimiter
	now := time.Unix(120, 0)
	for i := 0; i < 2; i++ {
		release, ok := limiter.acquire(7, assistantGenerationRequest, limits, now)
		require.True(t, ok)
		release()
	}
	_, ok := limiter.acquire(7, assistantGenerationRequest, limits, now)
	require.False(t, ok)
	release, ok := limiter.acquire(7, assistantStatusRequest, limits, now)
	require.True(t, ok)
	release()
	first, ok := limiter.acquire(7, assistantSearchExportRequest, limits, now)
	require.True(t, ok)
	_, ok = limiter.acquire(8, assistantSearchExportRequest, limits, now)
	require.False(t, ok)
	first()
	first()
	release, ok = limiter.acquire(8, assistantSearchExportRequest, limits, now)
	require.True(t, ok)
	release()
	_, ok = limiter.acquire(9, assistantStatusRequest, limits, now)
	require.False(t, ok)
	require.Len(t, limiter.users, 2)
	release, ok = limiter.acquire(9, assistantStatusRequest, limits, now.Add(time.Minute))
	require.True(t, ok)
	release()
	require.LessOrEqual(t, len(limiter.users), 2)
	release, ok = limiter.acquire(7, assistantGenerationRequest, limits, now.Add(time.Minute))
	require.True(t, ok)
	release()
}
