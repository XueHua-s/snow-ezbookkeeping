package api

import (
	"context"
	"errors"
	"github.com/mayswind/ezbookkeeping/pkg/core"
	"github.com/mayswind/ezbookkeeping/pkg/errs"
	"github.com/mayswind/ezbookkeeping/pkg/llm"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
	"math/big"
	"strconv"
	"strings"
	"time"
)

func assistantContinuation(c *data.ToolContinuation) *models.AIAssistantContinuation {
	if c == nil {
		return nil
	}
	return &models.AIAssistantContinuation{Tool: string(c.Tool), Cursor: c.Cursor}
}

func assistantQueries(queries []data.QueryEvidence) []models.AIAssistantQuery {
	result := make([]models.AIAssistantQuery, 0, len(queries))
	for _, q := range queries {
		s := q.Scope
		scope := models.AIAssistantQueryScope{StartTime: s.StartTime, EndTime: s.EndTime, Timezone: s.Timezone, Currency: s.Currency, Types: s.Types, AccountIDs: s.AccountIDs, CategoryIDs: s.CategoryIDs, TagIDs: s.TagIDs, TagMatch: s.TagMatch, NoTags: s.NoTags, TransferSide: s.TransferSide, AmountMin: s.AmountMin, AmountMax: s.AmountMax, Keyword: s.Keyword, IncludesHidden: s.IncludesHidden, ExcludesDeleted: s.ExcludesDeleted}
		for _, p := range s.Periods {
			scope.Periods = append(scope.Periods, models.AIAssistantQueryPeriod{Name: p.Name, StartTime: p.StartTime, EndTime: p.EndTime, MatchedCount: p.MatchedCount})
		}
		warnings := q.Warnings
		if warnings == nil {
			warnings = []string{}
		}
		result = append(result, models.AIAssistantQuery{QueryID: q.QueryID, Tool: string(q.Tool), Scope: scope, Coverage: string(q.Coverage), QueriedAt: q.QueriedAt, Warnings: warnings})
	}
	return result
}

// Tool amounts are exact two-decimal strings. Do not use float64 or ParseAmount
// (whose multiplication can overflow) at the legacy minor-unit boundary.
func assistantMinorAmount(amount string) (int64, error) {
	negative := strings.HasPrefix(amount, "-")
	digits := strings.TrimPrefix(amount, "-")
	parts := strings.Split(digits, ".")
	if len(parts) != 2 || len(parts[0]) == 0 || len(parts[1]) != 2 {
		return 0, errs.ErrAIAssistantFailed
	}
	for _, r := range parts[0] + parts[1] {
		if r < '0' || r > '9' {
			return 0, errs.ErrAIAssistantFailed
		}
	}
	number, ok := new(big.Int).SetString(parts[0]+parts[1], 10)
	if !ok {
		return 0, errs.ErrAIAssistantFailed
	}
	if negative {
		number.Neg(number)
	}
	if !number.IsInt64() {
		return 0, errs.ErrAIAssistantFailed
	}
	return number.Int64(), nil
}

func assistantReferences(refs []data.TransactionReference) ([]*models.AIAssistantReferencedTransaction, error) {
	result := make([]*models.AIAssistantReferencedTransaction, 0, len(refs))
	for _, ref := range refs {
		id, err := strconv.ParseInt(ref.ID, 10, 64)
		if err != nil || id <= 0 {
			return nil, errs.ErrAIAssistantFailed
		}
		at, err := time.Parse(time.RFC3339, ref.Time)
		if err != nil {
			return nil, errs.ErrAIAssistantFailed
		}
		amount, err := assistantMinorAmount(ref.SourceAmount)
		if err != nil {
			return nil, err
		}
		var destination int64
		if ref.DestinationAmount != "" {
			destination, err = assistantMinorAmount(ref.DestinationAmount)
			if err != nil {
				return nil, err
			}
		}
		var kind models.TransactionType
		switch ref.Type {
		case "income":
			kind = models.TRANSACTION_TYPE_INCOME
		case "expense":
			kind = models.TRANSACTION_TYPE_EXPENSE
		case "transfer":
			kind = models.TRANSACTION_TYPE_TRANSFER
		case "balance_modification":
			kind = models.TRANSACTION_TYPE_MODIFY_BALANCE
		default:
			return nil, errs.ErrAIAssistantFailed
		}
		result = append(result, &models.AIAssistantReferencedTransaction{Id: id, Time: at.Unix(), TimeText: at.In(time.FixedZone("", int(ref.TimezoneOffset)*60)).Format("2006-01-02 15:04:05"), Type: kind, CategoryName: ref.CategoryName, SourceAccountName: ref.SourceAccountName, DestinationAccountName: ref.DestinationAccountName, SourceAmount: amount, DestinationAmount: destination, Currency: ref.Currency, DestinationCurrency: ref.DestinationCurrency, Comment: ref.Comment, HideAmount: ref.HideAmount})
	}
	return result, nil
}

func assistantPreparationError(err *errs.Error) *errs.Error {
	if err.Context != nil {
		return err
	}
	return errs.NewErrorWithContext(err, core.O{"assistant": models.AIAssistantErrorContext{State: models.AIAssistantStateFailed, Queries: []models.AIAssistantQuery{}}})
}

func assistantRunError(result llm.AssistantRunResult, err error) *errs.Error {
	base := errs.ErrAIAssistantFailed
	var toolErr *data.ToolError
	if errors.As(err, &toolErr) {
		switch toolErr.Code {
		case data.ToolErrorInvalidArgument:
			base = errs.ErrAIAssistantInvalidArgument
		case data.ToolErrorInvalidCursor:
			base = errs.ErrAIAssistantInvalidCursor
		case data.ToolErrorNotFoundOrNotAccessible:
			base = errs.ErrAIAssistantNotAccessible
		case data.ToolErrorProviderUnsupported:
			base = errs.ErrAIAssistantProviderUnsupported
		case data.ToolErrorBudgetExhausted, data.ToolErrorResultTooLarge:
			base = errs.ErrAIAssistantIncomplete
		case data.ToolErrorQueryTimeout:
			base = errs.ErrAIAssistantTimeout
		}
	}
	state := models.AIAssistantStateFailed
	if result.State == llm.AssistantStateIncomplete || errors.Is(err, context.DeadlineExceeded) {
		state = models.AIAssistantStateIncomplete
	}
	if errors.Is(err, context.DeadlineExceeded) {
		base = errs.ErrAIAssistantTimeout
	}
	if errors.Is(err, context.Canceled) {
		base = errs.ErrAIAssistantCanceled
	}
	return errs.NewErrorWithContext(base, core.O{"assistant": models.AIAssistantErrorContext{State: state, Queries: assistantQueries(result.Queries), Continuation: assistantContinuation(result.Continuation)}})
}
