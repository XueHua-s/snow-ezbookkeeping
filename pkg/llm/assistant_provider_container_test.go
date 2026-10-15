package llm

import (
	"errors"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/llm/provider"
	"github.com/stretchr/testify/require"
)

type containerToolProvider struct {
	fallbackTestProvider
	turns int
}

func (p *containerToolProvider) RunToolTurn(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
	p.turns++
	return data.TurnResponse{}, errors.New("upstream")
}
func TestAssistantContainerExposesToolsWithoutFallbackPolicy(t *testing.T) {
	primary := &containerToolProvider{}
	fallback := &containerToolProvider{}
	container := &LargeLanguageModelProviderContainer{aiAssistantProvider: &fallbackLargeLanguageModelProvider{primaryProvider: primary, fallbackProvider: fallback}}
	p, f, err := container.GetAIAssistantToolProviders()
	require.NoError(t, err)
	require.Same(t, primary, p)
	require.Same(t, fallback, f)
	_, err = p.RunToolTurn(core.NewNullContext(), 1, data.ToolTurnInput{}, nil)
	require.Error(t, err)
	require.Equal(t, 1, primary.turns)
	require.Zero(t, fallback.turns)
	for _, configured := range []provider.LargeLanguageModelProvider{nil, &fallbackTestProvider{}, &fallbackLargeLanguageModelProvider{primaryProvider: primary, fallbackProvider: &fallbackTestProvider{}}} {
		container.aiAssistantProvider = configured
		p, f, err = container.GetAIAssistantToolProviders()
		require.Nil(t, p)
		require.Nil(t, f)
		var controlled *data.ToolError
		require.ErrorAs(t, err, &controlled)
		require.Equal(t, data.ToolErrorProviderUnsupported, controlled.Code)
	}
	container.aiAssistantProvider = primary
	p, f, err = container.GetAIAssistantToolProviders()
	require.NoError(t, err)
	require.Same(t, primary, p)
	require.Nil(t, f)
}
