package llm

import (
	"errors"
	"testing"

	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/stretchr/testify/require"
)

func TestRunnerProvisionalReplyRetractionAndNoDuplication(t *testing.T) {
	rounds := 0
	model := runnerModel(func(_ core.Context, _ int64, _ data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
		rounds++
		if rounds == 1 {
			require.NoError(t, emit(data.DisplayEvent{ReplyDelta: "unverified preamble"}))
			require.NoError(t, emit(data.DisplayEvent{ReplyReset: true}))
			require.NoError(t, emit(data.DisplayEvent{ReplyDelta: "must remain hidden"}))
			return data.TurnResponse{Completed: true, Content: "unverified preamble", ToolCalls: []data.ToolCall{runnerCall("q1", `{}`)}}, nil
		}
		require.NoError(t, emit(data.DisplayEvent{ReplyDelta: "verified "}))
		return data.TurnResponse{Completed: true, Content: "verified answer"}, nil
	})
	var displayed string
	resets := 0
	result, err := NewAssistantRunner(model, nil, &runnerExecutor{}).Run(runnerContext(), runnerInput(), func(e AssistantEvent) error {
		if e.Type == AssistantEventReplyDelta {
			displayed += e.Delta
		}
		if e.Type == AssistantEventReplyReset {
			displayed = ""
			resets++
		}
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, "verified answer", result.Reply)
	require.Equal(t, result.Reply, displayed)
	require.Equal(t, 1, resets)
}

func TestRunnerPartialSurvivesBrokenStreamWithoutFallback(t *testing.T) {
	model := runnerModel(func(_ core.Context, _ int64, _ data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
		require.NoError(t, emit(data.DisplayEvent{ReplyDelta: "partial"}))
		return data.TurnResponse{}, assistantRunnerError(data.ToolErrorProtocol)
	})
	fallback := runnerModel(func(core.Context, int64, data.ToolTurnInput, data.DisplayEventSink) (data.TurnResponse, error) {
		t.Fatal("visible reply must prohibit fallback")
		return data.TurnResponse{}, nil
	})
	result, err := NewAssistantRunner(model, fallback, &runnerExecutor{}).Run(runnerContext(), runnerInput(), func(AssistantEvent) error { return nil })
	require.Error(t, err)
	require.Equal(t, "partial", result.Reply)
	require.NotEqual(t, AssistantStateCompleted, result.State)
}

func TestRunnerReplySinkFailureStopsProvider(t *testing.T) {
	writeErr := errors.New("disconnected")
	model := runnerModel(func(_ core.Context, _ int64, _ data.ToolTurnInput, emit data.DisplayEventSink) (data.TurnResponse, error) {
		err := emit(data.DisplayEvent{ReplyDelta: "partial"})
		require.ErrorIs(t, err, writeErr)
		return data.TurnResponse{}, err
	})
	_, err := NewAssistantRunner(model, nil, &runnerExecutor{}).Run(runnerContext(), runnerInput(), func(e AssistantEvent) error {
		if e.Type == AssistantEventReplyDelta {
			return writeErr
		}
		return nil
	})
	require.ErrorIs(t, err, writeErr)
}
