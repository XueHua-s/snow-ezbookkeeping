package services

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/settings"
)

// AssistantRunPolicy freezes acceptance-time execution limits alongside capacity.
// Runtime is embedded so context preparation can decode the existing runtime DTO.
type AssistantRunPolicy struct {
	settings.AIAssistantRuntimeConfig
	Experience settings.AIAssistantExperienceConfig
}

func NewAssistantRunPolicy(runtime settings.AIAssistantRuntimeConfig, experience settings.AIAssistantExperienceConfig) AssistantRunPolicy {
	return AssistantRunPolicy{AIAssistantRuntimeConfig: runtime, Experience: experience}
}
func (s *AssistantRunService) policy(run *models.AssistantMessage) AssistantRunPolicy {
	p := NewAssistantRunPolicy(s.options.Runtime(), s.options.Config())
	// Empty policy is only for rows accepted before runtime snapshots were added.
	if run.RuntimeJSON != "" {
		_ = json.Unmarshal([]byte(run.RuntimeJSON), &p)
	}
	return p
}
