package llm

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"io"
)

func assistantCanonicalArguments(raw []byte) ([]byte, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	var value func() (any, error)
	value = func() (any, error) {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return token, nil
		}
		switch delimiter {
		case '{':
			object := map[string]any{}
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return nil, err
				}
				key, ok := keyToken.(string)
				if !ok {
					return nil, io.ErrUnexpectedEOF
				}
				if _, exists := object[key]; exists {
					return nil, io.ErrUnexpectedEOF
				}
				item, err := value()
				if err != nil {
					return nil, err
				}
				object[key] = item
			}
			_, err = decoder.Token()
			return object, err
		case '[':
			array := []any{}
			for decoder.More() {
				item, err := value()
				if err != nil {
					return nil, err
				}
				array = append(array, item)
			}
			_, err = decoder.Token()
			return array, err
		default:
			return nil, io.ErrUnexpectedEOF
		}
	}
	result, err := value()
	if err != nil {
		return nil, err
	}
	if _, err = decoder.Token(); err != io.EOF {
		return nil, io.ErrUnexpectedEOF
	}
	return json.Marshal(result)
}

func assistantKnownTool(name data.ToolName) bool {
	switch name {
	case data.ToolListAccounts, data.ToolListCategories, data.ToolListTags, data.ToolQueryTransactions, data.ToolGetTransaction, data.ToolAggregateTransactions, data.ToolGetAccountBalances:
		return true
	}
	return false
}

func assistantErrorCode(err error) data.ToolErrorCode {
	var controlled *data.ToolError
	if errors.As(err, &controlled) {
		switch controlled.Code {
		case data.ToolErrorInvalidArgument, data.ToolErrorNotFoundOrNotAccessible, data.ToolErrorQueryTimeout, data.ToolErrorCanceled, data.ToolErrorBudgetExhausted, data.ToolErrorProviderUnsupported, data.ToolErrorUpstreamFailed, data.ToolErrorResultTooLarge, data.ToolErrorInvalidCursor, data.ToolErrorDataInconsistent, data.ToolErrorProtocol:
			return controlled.Code
		}
	}
	return data.ToolErrorUpstreamFailed
}

func assistantRunnerError(code data.ToolErrorCode) *data.ToolError {
	return &data.ToolError{Code: code, Message: "The assistant could not complete this request.", Retryable: code == data.ToolErrorQueryTimeout || code == data.ToolErrorUpstreamFailed}
}

func assistantReason(code data.ToolErrorCode) AssistantStopReason {
	switch code {
	case data.ToolErrorInvalidArgument:
		return AssistantReasonInvalidArgument
	case data.ToolErrorNotFoundOrNotAccessible:
		return AssistantReasonNotAccessible
	case data.ToolErrorQueryTimeout:
		return AssistantReasonQueryTimeout
	case data.ToolErrorCanceled:
		return AssistantReasonCanceled
	case data.ToolErrorBudgetExhausted:
		return AssistantReasonContextBudget
	case data.ToolErrorProviderUnsupported:
		return AssistantReasonProviderUnsupported
	case data.ToolErrorResultTooLarge:
		return AssistantReasonResultBudget
	case data.ToolErrorInvalidCursor:
		return AssistantReasonInvalidCursor
	case data.ToolErrorDataInconsistent:
		return AssistantReasonDataInconsistent
	case data.ToolErrorProtocol:
		return AssistantReasonProtocolError
	default:
		return AssistantReasonUpstreamFailed
	}
}
