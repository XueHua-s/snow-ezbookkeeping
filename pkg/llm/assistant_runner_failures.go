package llm

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
)

type assistantUnresolvedCall struct {
	call    data.ToolCall
	failure data.ToolError
	round   int
}

// An exact retry takes precedence. Otherwise only a unique, explicit parameter
// correction can advance a failed query; one success cannot erase several calls.
func (run *assistantRun) resolveFailure(call data.ToolCall, key [32]byte) {
	if _, exists := run.unresolved[key]; exists {
		delete(run.unresolved, key)
		return
	}
	run.advanceArgumentCorrection(call)
}

func (run *assistantRun) recordFailure(call data.ToolCall, key [32]byte, failure data.ToolError) {
	if previous, exists := run.unresolved[key]; exists {
		// A later validation error must not weaken an existing read/authority
		// failure into something that can be cleared by changing arguments.
		if previous.failure.Code != data.ToolErrorInvalidArgument {
			return
		}
	} else {
		// Corrected validation can expose another validation or a read failure.
		// Transfer the unique ancestor, but retain the new failure under the
		// corrected identity so read/authority failures require an exact retry.
		run.advanceArgumentCorrection(call)
	}
	run.unresolved[key] = assistantUnresolvedCall{call: call, failure: failure, round: run.toolRound}
}

func (run *assistantRun) advanceArgumentCorrection(call data.ToolCall) {
	var candidate [32]byte
	matches := 0
	for key, failed := range run.unresolved {
		if failed.round < run.toolRound && assistantCorrectsArgument(failed, call) {
			candidate = key
			matches++
		}
	}
	if matches == 1 {
		delete(run.unresolved, candidate)
	}
}

// A validation correction may change only the explicitly rejected parameter.
// Unlabelled or array-element validation failures remain unresolved: the runner cannot
// infer query intent or safely equate a different financial scope with a retry.
func assistantCorrectsArgument(failed assistantUnresolvedCall, call data.ToolCall) bool {
	field := strings.TrimPrefix(failed.failure.Field, "arguments.")
	if failed.call.Name != call.Name || failed.failure.Code != data.ToolErrorInvalidArgument ||
		field == "" || field == "arguments" || field == "duplicate_key" || strings.ContainsAny(field, "[]") {
		return false
	}
	withoutField := func(raw []byte) ([]byte, bool) {
		canonical, err := assistantCanonicalArguments(raw)
		if err != nil {
			return nil, false
		}
		var arguments map[string]json.RawMessage
		if json.Unmarshal(canonical, &arguments) != nil || arguments == nil {
			return nil, false
		}
		var remove func(map[string]json.RawMessage, []string) bool
		remove = func(object map[string]json.RawMessage, path []string) bool {
			if path[0] == "" {
				return false
			}
			if len(path) == 1 {
				delete(object, path[0])
				return true
			}
			var child map[string]json.RawMessage
			if json.Unmarshal(object[path[0]], &child) != nil || child == nil || !remove(child, path[1:]) {
				return false
			}
			object[path[0]], _ = json.Marshal(child)
			return true
		}
		if !remove(arguments, strings.Split(field, ".")) {
			return nil, false
		}
		encoded, err := json.Marshal(arguments)
		return encoded, err == nil
	}
	before, validBefore := withoutField(failed.call.Arguments)
	after, validAfter := withoutField(call.Arguments)
	return validBefore && validAfter && bytes.Equal(before, after)
}
