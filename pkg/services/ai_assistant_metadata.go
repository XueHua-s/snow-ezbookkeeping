package services

import (
	"sort"
	"strconv"

	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"github.com/mayswind/ezbookkeeping/pkg/models"
)

func newAssistantMetadataTools(s *AssistantQueryService) map[data.ToolName]assistantToolHandler {
	return map[data.ToolName]assistantToolHandler{
		data.ToolListAccounts:       assistantListAccounts,
		data.ToolListCategories:     assistantListCategories,
		data.ToolListTags:           assistantListTags,
		data.ToolGetAccountBalances: assistantAccountBalances,
	}
}

type assistantAccountMetadata struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	ParentID string                 `json:"parent_id"`
	Type     models.AccountType     `json:"type"`
	Category models.AccountCategory `json:"category"`
	Currency string                 `json:"currency"`
	Hidden   bool                   `json:"hidden"`
}
type assistantCategoryMetadata struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	ParentID string `json:"parent_id"`
	Type     string `json:"type"`
	Hidden   bool   `json:"hidden"`
}
type assistantTagMetadata struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	TagGroupID   string `json:"tag_group_id"`
	TagGroupName string `json:"tag_group_name"`
	Hidden       bool   `json:"hidden"`
}

func assistantListAccounts(e *assistantQueryExecution) (data.ToolResult, error) {
	rows := make([]assistantEntityRow, 0, len(e.Metadata.Accounts))
	for id, a := range e.Metadata.Accounts {
		if assistantLiteralContains(a.Name, e.Args.Keyword) {
			rows = append(rows, assistantEntityRow{id, assistantAccountMetadata{assistantCount(id), a.Name, assistantCount(a.ParentAccountId), a.Type, a.Category, a.Currency, a.Hidden}})
		}
	}
	return assistantEntityPage(e, rows, data.ToolResult{})
}
func assistantCategoryType(t models.TransactionCategoryType) string {
	switch t {
	case models.CATEGORY_TYPE_INCOME:
		return "income"
	case models.CATEGORY_TYPE_EXPENSE:
		return "expense"
	case models.CATEGORY_TYPE_TRANSFER:
		return "transfer"
	default:
		return "unknown"
	}
}
func assistantListCategories(e *assistantQueryExecution) (data.ToolResult, error) {
	rows := make([]assistantEntityRow, 0, len(e.Metadata.Categories))
	for id, c := range e.Metadata.Categories {
		kind := assistantCategoryType(c.Type)
		if (e.Args.Type == "" || e.Args.Type == kind) && assistantLiteralContains(c.Name, e.Args.Keyword) {
			rows = append(rows, assistantEntityRow{id, assistantCategoryMetadata{assistantCount(id), c.Name, assistantCount(c.ParentCategoryId), kind, c.Hidden}})
		}
	}
	return assistantEntityPage(e, rows, data.ToolResult{})
}
func assistantListTags(e *assistantQueryExecution) (data.ToolResult, error) {
	rows := make([]assistantEntityRow, 0, len(e.Metadata.Tags))
	for id, tag := range e.Metadata.Tags {
		if assistantLiteralContains(tag.Name, e.Args.Keyword) {
			groupName := ""
			if group := e.Metadata.TagGroups[tag.TagGroupId]; group != nil {
				groupName = group.Name
			}
			rows = append(rows, assistantEntityRow{id, assistantTagMetadata{assistantCount(id), tag.Name, assistantCount(tag.TagGroupId), groupName, tag.Hidden}})
		}
	}
	return assistantEntityPage(e, rows, data.ToolResult{})
}

type assistantEntityRow struct {
	id    int64
	value any
}

// Entity pages share numerical ID ordering and keep the cursor at the last
// emitted row even when the complete envelope forces a shorter page. Balances
// supply full-range totals before paging; the high-water bounds only details.
func assistantEntityPage(e *assistantQueryExecution, rows []assistantEntityRow, result data.ToolResult) (data.ToolResult, error) {
	sort.Slice(rows, func(i, j int) bool { return rows[i].id < rows[j].id })
	result.OK = true
	result.Coverage = data.QueryCoveragePage
	result.MatchedCount = strconv.Itoa(len(rows))
	result.Data = assistantJSON([]any{})
	high, last := int64(0), int64(0)
	if len(rows) > 0 {
		high = rows[len(rows)-1].id
	}
	if e.Cursor != nil {
		var err error
		high, err = assistantID(e.Cursor.HighWater.ID)
		if err != nil {
			return data.ToolResult{}, assistantCursorError()
		}
		last, err = assistantID(e.Cursor.Last.ID)
		if err != nil {
			return data.ToolResult{}, assistantCursorError()
		}
	}
	start := sort.Search(len(rows), func(i int) bool { return rows[i].id > last })
	end := sort.Search(len(rows), func(i int) bool { return rows[i].id > high })
	values := make([]any, 0, e.Args.PageSize)
	for i := start; i < end && len(values) < e.Args.PageSize; i++ {
		if err := e.Context.Err(); err != nil {
			return data.ToolResult{}, err
		}
		candidate := result
		candidateValues := append(values, rows[i].value)
		candidate.Data = assistantJSON(candidateValues)
		candidate.ReturnedCount = len(candidateValues)
		candidate.HasMore = i+1 < end
		candidate.NextCursor = ""
		if candidate.HasMore {
			var err error
			candidate.NextCursor, err = e.NextCursor(assistantPageKey{ID: assistantCount(high)}, assistantPageKey{ID: assistantCount(rows[i].id)})
			if err != nil {
				return data.ToolResult{}, err
			}
		}
		if !e.PageFits(candidate) {
			if len(values) == 0 {
				return data.ToolResult{}, &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "An entity and its query envelope exceed the result limit"}
			}
			break
		}
		values = candidateValues
		result = candidate
	}
	if !e.PageFits(result) {
		return data.ToolResult{}, &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "The query envelope exceeds the result limit"}
	}
	return result, nil
}
