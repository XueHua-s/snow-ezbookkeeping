package api

import (
	"context"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/datastore"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/log"
	"github.com/mayswind/ezbookkeeping/pkg/services"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
	"sync"
	"time"
)

// InitializeAssistantExperience runs once during application startup, after
// configuration and datastores are ready and before requests or workers start.
func (a *LargeLanguageModelsApi) InitializeAssistantExperience() {
	if a.experience != nil {
		return
	}
	storage := services.NewAssistantStorageService(datastore.Container)
	config := func() settings.AIAssistantExperienceConfig { return a.CurrentConfig().GetAIAssistantExperienceConfig() }
	storage.SetConfigProvider(config)
	DataManagements.SetAssistantStorage(storage)
	services.Users.SetAssistantStorage(storage)
	available := func(ctx core.Context, uid int64) bool {
		cfg := a.CurrentConfig()
		if !cfg.EnableAIAssistant || cfg.AIAssistantLLMConfig == nil || cfg.AIAssistantLLMConfig.LLMProvider == "" {
			return false
		}
		user, err := a.users.GetUserById(ctx, uid)
		if err != nil || user.Disabled {
			return false
		}
		_, _, err = llm.Container.GetAIAssistantToolProviders()
		return err == nil
	}
	contextService := services.NewAssistantContextService(storage, a.CurrentConfig)
	a.experience = &assistantExperienceDependencies{
		storage:             storage,
		cleanup:             services.NewAssistantCleanupService(storage),
		conversations:       services.NewAssistantConversationService(storage, services.AssistantConversationOptions{Config: config, Runtime: func() settings.AIAssistantRuntimeConfig { return a.CurrentConfig().GetAIAssistantRuntimeConfig() }, Secret: []byte(a.CurrentConfig().SecretKey), GenerationAvailable: available}),
		memories:            services.NewAssistantMemoryService(storage, a.CurrentConfig),
		runs:                services.NewAssistantRunService(storage, services.AssistantRunOptions{Config: config, Runtime: func() settings.AIAssistantRuntimeConfig { return a.CurrentConfig().GetAIAssistantRuntimeConfig() }, Prepare: contextService.Prepare, GenerationAvailable: available}),
		generationAvailable: available,
	}
}
func (a *LargeLanguageModelsApi) StartAssistantExperience(parent context.Context) {
	if a.experience == nil {
		return
	}
	d := a.experience
	d.lifecycleMu.Lock()
	defer d.lifecycleMu.Unlock()
	if d.backgroundCancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	d.backgroundCancel = cancel
	d.backgroundDone = make(chan struct{})
	d.runs.Start(ctx)
	var tasks sync.WaitGroup
	for _, sweep := range []func(core.Context) error{d.cleanup.Sweep, d.memories.Sweep} {
		tasks.Add(1)
		go func(work func(core.Context) error) {
			defer tasks.Done()
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					job := core.NewAssistantContext(ctx, "assistant-maintenance", "", "")
					if err := work(job); err != nil && ctx.Err() == nil {
						log.Warnf(job, "[assistant] background maintenance failed")
					}
				}
			}
		}(sweep)
	}
	go func() { tasks.Wait(); close(d.backgroundDone) }()
}
func (a *LargeLanguageModelsApi) StopAssistantExperience(ctx context.Context) error {
	if a.experience == nil {
		return nil
	}
	d := a.experience
	d.lifecycleMu.Lock()
	cancel, done := d.backgroundCancel, d.backgroundDone
	d.lifecycleMu.Unlock()
	if cancel != nil {
		cancel()
	}
	if err := d.runs.Stop(ctx); err != nil {
		return err
	}
	if done != nil {
		select {
		case <-done:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}
