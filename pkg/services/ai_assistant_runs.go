package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// AssistantRunPreparer constructs bounded server-authoritative context while the
// run is preparing. Its context belongs to the application, never an HTTP client.
type AssistantRunPreparer func(core.Context, *models.AssistantMessage) (*AssistantPreparedRun, error)
type AssistantPreparedRun struct {
	Runner       *llm.AssistantRunner
	Input        llm.AssistantRunInput
	Dependencies models.AssistantInputDependencies
}
type AssistantRunObservation struct {
	RunId               int64
	State               models.AssistantRunState
	Reason              models.AssistantReasonCode
	Revision            int64
	ElapsedMilliseconds int64
}
type AssistantRunOptions struct {
	Observe             func(core.Context, AssistantRunObservation)
	Config              func() settings.AIAssistantExperienceConfig
	Runtime             func() settings.AIAssistantRuntimeConfig
	Prepare             AssistantRunPreparer
	GenerationAvailable func(core.Context, int64) bool
}
type AssistantRunService struct {
	storage   *AssistantStorageService
	options   AssistantRunOptions
	owner     string
	mu        sync.Mutex
	cancel    context.CancelFunc
	done      chan struct{}
	slots     chan struct{}
	scanAfter map[int]int64
}

func NewAssistantRunService(storage *AssistantStorageService, options AssistantRunOptions) *AssistantRunService {
	if options.Observe == nil {
		options.Observe = func(ctx core.Context, o AssistantRunObservation) {
			log.Infof(ctx, "[assistant-run] run=%d state=%s reason=%s revision=%d elapsed_ms=%d", o.RunId, o.State, o.Reason, o.Revision, o.ElapsedMilliseconds)
		}
	}
	if options.Runtime == nil {
		options.Runtime = (&settings.Config{}).GetAIAssistantRuntimeConfig
	}
	if options.Config == nil {
		options.Config = settings.DefaultAIAssistantExperienceConfig
	}
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return &AssistantRunService{storage: storage, options: options, owner: hex.EncodeToString(b)}
}

// Start is idempotent. Each scan is bounded; only claimed runs occupy goroutines.
func (s *AssistantRunService) Start(parent context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel
	s.done = make(chan struct{})
	s.slots = make(chan struct{}, s.options.Config().Limits.WorkerConcurrency)
	s.scanAfter = make(map[int]int64)
	go s.loop(ctx)
}
func (s *AssistantRunService) Stop(ctx context.Context) error {
	s.mu.Lock()
	cancel, done := s.cancel, s.done
	s.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func (s *AssistantRunService) loop(ctx context.Context) {
	defer close(s.done)
	var work sync.WaitGroup
	defer work.Wait()
	ticker := time.NewTicker(time.Duration(s.options.Config().CheckpointMilliseconds) * time.Millisecond)
	defer ticker.Stop()
	for {
		if ctx.Err() != nil {
			return
		}
		s.scan(ctx, &work)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (s *AssistantRunService) observe(ctx core.Context, run *models.AssistantMessage) {
	s.options.Observe(ctx, AssistantRunObservation{RunId: run.Id, State: run.RunState, Reason: run.ReasonCode, Revision: run.SnapshotRevision, ElapsedMilliseconds: s.storage.now().UnixMilli() - run.CreatedAt})
}
