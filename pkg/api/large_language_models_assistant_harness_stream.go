package api

import (
	"context"
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"github.com/mayswind/ezbookkeeping/pkg/utils"
	"sync"
	"time"
)

type assistantStreamResult struct {
	result llm.AssistantRunResult
	err    error
}

func (a *LargeLanguageModelsApi) AssistantChatStreamHandler(c *core.WebContext) *errs.Error {
	request, err := a.prepareAssistantRun(c)
	if err != nil {
		return assistantPreparationError(err)
	}
	defer request.cancel()
	utils.SetEventStreamHeader(c)
	socketContext := c.Request.Context()
	events := make(chan assistantStreamEvent)
	done := make(chan struct{})
	writesStopped := make(chan struct{})
	var firstMu sync.Mutex
	var first *time.Time
	request.input.FirstStatusWrittenAt = func() *time.Time {
		firstMu.Lock()
		defer firstMu.Unlock()
		if first == nil {
			return nil
		}
		value := *first
		return &value
	}
	request.input.Finalize = func(result llm.AssistantRunResult, runErr error) error {
		if socketContext.Err() != nil || errors.Is(context.Cause(request.context), llm.ErrAssistantWriteFailed) {
			return nil
		}
		delivery := assistantStreamEvent{terminal: &assistantStreamResult{result, runErr}, ack: make(chan error, 1)}
		select {
		case events <- delivery:
			return <-delivery.ack
		case <-socketContext.Done():
			return nil
		case <-writesStopped:
			return nil
		}
	}
	go func() {
		defer close(done)
		_, _ = request.runner.Run(request.context, request.input, func(event llm.AssistantEvent) error {
			delivery := assistantStreamEvent{event: event, ack: make(chan error, 1)}
			select {
			case events <- delivery:
			case <-request.context.Done():
				return request.context.Err()
			}
			// Every delivered event is acknowledged by the sole writer, even on failure.
			return <-delivery.ack
		})
	}()
	ticker := time.NewTicker(request.heartbeat)
	defer ticker.Stop()
	stop := func() {
		close(writesStopped)
		request.cancelCause(llm.ErrAssistantWriteFailed)
		request.cancel()
		<-done
	}
	for {
		select {
		case delivery := <-events:
			var writeErr error
			if delivery.terminal != nil {
				writeErr = writeAssistantTerminal(c, *delivery.terminal)
				// No heartbeat or second frame can follow a terminal, including a failed
				// terminal write. Its acknowledgment lets the runner finalize telemetry.
				ticker.Stop()
				delivery.ack <- writeErr
				if writeErr != nil {
					stop()
				} else {
					<-done
				}
				return nil
			}
			event := delivery.event
			refs, projectionErr := assistantReferences(event.References)
			writeErr = projectionErr
			if writeErr == nil {
				writeErr = utils.WriteEventStreamJSON(c, &models.AIAssistantChatStreamChunk{Type: string(event.Type), Status: models.AIAssistantStatus(event.Status), Delta: event.Delta, References: refs})
			}
			if writeErr == nil && event.Type == llm.AssistantEventStatus {
				firstMu.Lock()
				if first == nil {
					at := request.now()
					first = &at
				}
				firstMu.Unlock()
			}
			delivery.ack <- writeErr
			if writeErr != nil {
				stop()
				return nil
			}
		case <-ticker.C:
			if writeErr := utils.WriteEventStreamComment(c, "heartbeat"); writeErr != nil {
				stop()
				return nil
			}
		case <-socketContext.Done():
			close(writesStopped)
			request.cancel()
			<-done
			return nil
		case <-done:
			return nil
		}
	}
}

func writeAssistantTerminal(c *core.WebContext, completed assistantStreamResult) error {
	result := completed.result
	refs, projectionErr := assistantReferences(result.References)
	if completed.err == nil {
		completed.err = projectionErr
	}
	if completed.err != nil {
		apiErr := assistantRunError(result, completed.err)
		c.SetResponseError(apiErr)
		envelope := utils.GetJsonErrorResult(apiErr, c.Request.URL.Path)
		envelope["context"] = apiErr.Context
		return utils.WriteEventStreamJSON(c, envelope)
	}
	return utils.WriteEventStreamJSON(c, &models.AIAssistantChatStreamChunk{Type: models.AIAssistantChatStreamChunkTypeDone, Mode: result.Mode, Reply: result.Reply, Thinking: result.Thinking, References: refs, Queries: assistantQueries(result.Queries), Continuation: assistantContinuation(result.Continuation)})
}
