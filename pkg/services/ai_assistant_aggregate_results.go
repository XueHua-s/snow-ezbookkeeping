package services

import (
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"math/big"
	"sort"
	"time"
)

func (a *assistantAggregation) page() ([]*assistantAggregateGroup, error) {
	limit := a.e.Args.PageSize + 1
	var last *assistantAggregatePosition
	if a.e.Cursor != nil {
		last = &assistantAggregatePosition{}
		if err := json.Unmarshal([]byte(a.e.Cursor.Last.Group), last); err != nil {
			return nil, assistantCursorError()
		}
	}
	page := make([]*assistantAggregateGroup, 0, limit)
	offer := func(g *assistantAggregateGroup) {
		pos := a.position(g)
		if last != nil && a.compare(pos, *last) <= 0 {
			return
		}
		i := sort.Search(len(page), func(i int) bool { return a.compare(a.position(page[i]), pos) > 0 })
		if i >= limit {
			return
		}
		page = append(page, nil)
		copy(page[i+1:], page[i:])
		page[i] = g
		if len(page) > limit {
			page = page[:limit]
		}
	}
	for _, g := range a.groups {
		if err := a.e.Context.Err(); err != nil {
			return nil, err
		}
		offer(g)
	}
	// Only observed non-time combinations participate. Each series contributes at
	// most page_size+1 empty candidates; binary search skips arbitrarily deep pages.
	for d := range a.combinations {
		for p := range a.ranges {
			count := a.bucketCount(p)
			if count == 0 {
				continue
			}
			reverse := a.e.Args.OrderBy.Field == "group_key" && a.e.Args.OrderBy.Direction == "desc"
			groupAt := func(i int) *assistantAggregateGroup {
				if reverse {
					i = count - 1 - i
				}
				return &assistantAggregateGroup{Period: p, Dimension: d, Ordinal: i}
			}
			start := 0
			if last != nil {
				start = sort.Search(count, func(i int) bool { return a.compare(a.position(groupAt(i)), *last) > 0 })
			}
			accepted := 0
			for i := start; i < count && accepted < limit; i++ {
				if err := a.e.Context.Err(); err != nil {
					return nil, err
				}
				g := groupAt(i)
				if !a.validBucket(p, g.Ordinal) {
					continue
				}
				if a.groups[a.storageKey(g)] != nil {
					continue
				}
				offer(g)
				accepted++
			}
		}
	}
	return page, nil
}

// Ratios are rounded in integer hundredths of a percent, including negative
// deltas. Neither database SUM nor binary floating point participates.
func assistantAggregatePercentage(delta, baseline *big.Int) (*string, string) {
	if baseline.Sign() <= 0 {
		return nil, "non_positive_baseline"
	}
	numerator := new(big.Int).Mul(delta, big.NewInt(10000))
	sign := numerator.Sign()
	numerator.Abs(numerator)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(numerator, baseline, remainder)
	if remainder.Lsh(remainder, 1).Cmp(baseline) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	if sign < 0 {
		quotient.Neg(quotient)
	}
	value := assistantFormatAmount(quotient)
	return &value, ""
}

func (a *assistantAggregation) comparison(g *assistantAggregateGroup) map[string]any {
	c := a.e.Args.Comparison
	if c == nil {
		return nil
	}
	baseline, current := -1, -1
	for p, r := range a.ranges {
		if r.Name == c.Baseline {
			baseline = p
		}
		if r.Name == c.Current {
			current = p
		}
	}
	if g.Period != baseline && g.Period != current {
		return nil
	}
	out := map[string]any{"group": a.project(g, false), "baseline_period": c.Baseline, "current_period": c.Current}
	values := map[string]any{}
	paired := a.validBucket(baseline, g.Ordinal) && a.validBucket(current, g.Ordinal)
	for _, metric := range a.e.Args.Metrics {
		pair := map[string]any{"baseline": nil, "current": nil, "delta": nil, "percentage": nil}
		numbers := make([]*big.Int, 2)
		for i, p := range []int{baseline, current} {
			if !a.validBucket(p, g.Ordinal) {
				continue
			}
			other := &assistantAggregateGroup{Period: p, Dimension: g.Dimension, Ordinal: g.Ordinal}
			if actual := a.groups[a.storageKey(other)]; actual != nil {
				other = actual
			}
			numbers[i] = other.Stats.metric(metric)
			if numbers[i] != nil {
				if metric == "count" {
					pair[[]string{"baseline", "current"}[i]] = numbers[i].String()
				} else {
					pair[[]string{"baseline", "current"}[i]] = assistantFormatAmount(numbers[i])
				}
			}
		}
		if !paired {
			pair["reason"] = "unmatched_time_bucket"
		} else if numbers[0] == nil || numbers[1] == nil {
			pair["reason"] = "empty_metric"
		} else {
			delta := new(big.Int).Sub(numbers[1], numbers[0])
			if metric == "count" {
				pair["delta"] = delta.String()
			} else {
				pair["delta"] = assistantFormatAmount(delta)
			}
			percentage, reason := assistantAggregatePercentage(delta, numbers[0])
			pair["percentage"] = percentage
			if reason != "" {
				pair["reason"] = reason
			}
		}
		values[metric] = pair
	}
	out["paired"] = paired
	out["metrics"] = values
	return out
}

func assistantAggregateTransactions(e *assistantQueryExecution) (data.ToolResult, error) {
	a := &assistantAggregation{e: e, groups: map[string]*assistantAggregateGroup{}, totals: map[string]*assistantAggregateGroup{}, combinations: map[assistantAggregateDimension]bool{}, additive: true}
	for _, dimension := range e.Args.GroupBy {
		if dimension == "day" || dimension == "month" || dimension == "year" {
			a.dimension = dimension
		}
	}
	if len(e.Args.Periods) > 0 {
		for _, p := range e.Args.Periods {
			start, end := time.UnixMilli(p.StartMillis).In(e.Service.location), time.UnixMilli(p.EndMillis).In(e.Service.location)
			a.ranges = append(a.ranges, assistantAggregateRange{Name: p.Name, Start: &start, End: &end})
		}
	} else {
		r := assistantAggregateRange{}
		if e.Args.Filter.StartMillis != nil {
			v := time.UnixMilli(*e.Args.Filter.StartMillis).In(e.Service.location)
			r.Start = &v
		}
		if e.Args.Filter.EndMillis != nil {
			v := time.UnixMilli(*e.Args.Filter.EndMillis).In(e.Service.location)
			r.End = &v
		}
		a.ranges = []assistantAggregateRange{r}
	}
	if err := a.scan(); err != nil {
		return data.ToolResult{}, err
	}
	if err := a.fillEmpty(); err != nil {
		return data.ToolResult{}, err
	}
	totals := make([]*assistantAggregateGroup, 0, len(a.totals))
	for _, g := range a.totals {
		totals = append(totals, g)
	}
	sort.Slice(totals, func(i, j int) bool { return a.storageKey(totals[i]) < a.storageKey(totals[j]) })
	projectedTotals := make([]map[string]any, 0, len(totals))
	for _, g := range totals {
		projectedTotals = append(projectedTotals, a.project(g, true))
	}
	result := data.ToolResult{OK: true, MatchedCount: assistantCount(a.matched), Coverage: data.QueryCoveragePage, TotalsCoverage: data.QueryCoverageAllMatching, Totals: assistantJSON(projectedTotals), CountsAdditive: &a.additive, Warnings: []string{}}
	for p, r := range a.ranges {
		if len(e.Args.Periods) > 0 {
			result.Periods = append(result.Periods, data.QueryPeriod{Name: r.Name, StartTime: e.Args.Periods[p].StartTime, EndTime: e.Args.Periods[p].EndTime, MatchedCount: assistantCount(r.Count)})
		}
	}
	page, err := a.page()
	if err != nil {
		return data.ToolResult{}, err
	}
	high, err := e.HighWater()
	if err != nil {
		return data.ToolResult{}, err
	}
	rows := make([]map[string]any, 0, e.Args.PageSize)
	comparisons := make([]map[string]any, 0, e.Args.PageSize)
	result.Data = assistantJSON(rows)
	more := false
	result.GroupsHasMore = &more
	if e.Args.Comparison != nil {
		result.Comparison = assistantJSON(comparisons)
	}
	for i, g := range page {
		if i >= e.Args.PageSize {
			break
		}
		nextRows := append(rows, a.project(g, false))
		nextComparisons := comparisons
		if comparison := a.comparison(g); comparison != nil {
			nextComparisons = append(comparisons, comparison)
		}
		candidate := result
		candidate.Data = assistantJSON(nextRows)
		candidate.ReturnedCount = len(nextRows)
		if e.Args.Comparison != nil {
			candidate.Comparison = assistantJSON(nextComparisons)
		}
		candidate.HasMore = i+1 < len(page)
		candidate.GroupsHasMore = &candidate.HasMore
		candidate.NextCursor = ""
		if candidate.HasMore {
			candidate.NextCursor, err = e.NextCursor(high, assistantPageKey{Group: string(assistantJSON(a.position(g)))})
			if err != nil {
				return data.ToolResult{}, err
			}
		}
		if !e.PageFits(candidate) {
			if len(rows) == 0 {
				return data.ToolResult{}, &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "Aggregate item and complete totals exceed the result limit"}
			}
			break
		}
		result = candidate
		rows = nextRows
		comparisons = nextComparisons
	}
	if !e.PageFits(result) {
		return data.ToolResult{}, &data.ToolError{Code: data.ToolErrorResultTooLarge, Message: "Complete aggregate totals exceed the result limit"}
	}
	return result, nil
}
