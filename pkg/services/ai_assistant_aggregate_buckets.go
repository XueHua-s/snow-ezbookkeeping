package services

import (
	"strconv"
	"time"
)

// RFC3339 cannot encode historical timezone offsets with seconds. Preserve
// the exact boundary instant by using UTC for those historical buckets.
func assistantAggregateTime(t time.Time) string {
	if _, offset := t.Zone(); offset%60 != 0 {
		t = t.UTC()
	}
	return t.Format(time.RFC3339)
}

func (a *assistantAggregation) project(g *assistantAggregateGroup, total bool) map[string]any {
	out := g.Stats.project(a.e.Args.Metrics)
	out["currency"] = g.Dimension.Currency
	out["flow"] = g.Dimension.Flow
	if len(a.e.Args.Periods) > 0 {
		out["period"] = a.ranges[g.Period].Name
	}
	if total {
		return out
	}
	if containsString(a.e.Args.GroupBy, "account") {
		out["account_id"] = g.Dimension.Account
		id, _ := strconv.ParseInt(g.Dimension.Account, 10, 64)
		out["account_name"] = a.e.Metadata.Accounts[id].Name
	}
	if containsString(a.e.Args.GroupBy, "category") {
		out["category_id"] = nil
		out["category_name"] = "Uncategorized"
		if g.Dimension.Category != "" {
			out["category_id"] = g.Dimension.Category
			id, _ := strconv.ParseInt(g.Dimension.Category, 10, 64)
			out["category_name"] = a.e.Metadata.Categories[id].Name
		}
	}
	if a.dimension != "" {
		start := a.bucketStart(g.Period, g.Ordinal)
		end := a.bucketStart(g.Period, g.Ordinal+1)
		r := a.ranges[g.Period]
		if r.Start != nil && start.Before(*r.Start) {
			start = *r.Start
		}
		if r.End != nil && end.After(*r.End) {
			end = *r.End
		}
		out["time_bucket"] = map[string]any{"unit": a.dimension, "start_time": assistantAggregateTime(start), "end_time": assistantAggregateTime(end), "ordinal": strconv.Itoa(g.Ordinal)}
	}
	return out
}

func (a *assistantAggregation) bucketStart(p, ordinal int) time.Time {
	if r := a.ranges[p]; r.Start != nil {
		ordinal += assistantCalendarIndex(*r.Start, a.dimension)
	}
	var civil time.Time
	switch a.dimension {
	case "year":
		civil = time.Date(ordinal, 1, 1, 0, 0, 0, 0, time.UTC)
	case "month":
		civil = time.Date(ordinal/12, time.Month(ordinal%12+1), 1, 0, 0, 0, 0, time.UTC)
	default:
		civil = time.Unix(int64(ordinal)*86400, 0).UTC()
	}
	return assistantCalendarBoundary(civil, a.e.Service.location)
}

func (a *assistantAggregation) validBucket(p, ordinal int) bool {
	return a.dimension == "" || ordinal < a.bucketCount(p) && a.bucketStart(p, ordinal).Before(a.bucketStart(p, ordinal+1))
}

func (a *assistantAggregation) scan() error {
	return a.e.ScanTransactions(nil, func(tx *assistantLogicalTransaction) error {
		matched := false
		members := 0
		for p := range a.ranges {
			r := &a.ranges[p]
			timestamp := time.UnixMilli(tx.Record.TransactionTime).In(a.e.Service.location)
			if r.Start != nil && timestamp.Before(*r.Start) || r.End != nil && !timestamp.Before(*r.End) {
				continue
			}
			matched = true
			members++
			r.Count++
			for _, side := range tx.MatchedSides {
				d := assistantAggregateDimension{Currency: side.Account.Currency, Flow: side.Flow}
				total := &assistantAggregateGroup{Period: p, Dimension: d}
				key := a.storageKey(total)
				if a.totals[key] == nil {
					a.totals[key] = total
				}
				a.totals[key].Stats.add(side.Amount)
				if containsString(a.e.Args.GroupBy, "account") {
					d.Account = strconv.FormatInt(side.Account.AccountId, 10)
				}
				if containsString(a.e.Args.GroupBy, "category") && tx.Category != nil {
					d.Category = strconv.FormatInt(tx.Category.CategoryId, 10)
				}
				g := &assistantAggregateGroup{Period: p, Dimension: d, Ordinal: a.ordinal(p, timestamp)}
				key = a.storageKey(g)
				if a.groups[key] == nil {
					a.groups[key] = g
				}
				a.groups[key].Stats.add(side.Amount)
				a.combinations[d] = true
			}
		}
		if matched {
			a.matched++
			if len(tx.MatchedSides) > 1 || members > 1 {
				a.additive = false
			}
		}
		return nil
	})
}

// Missing periods inherit only observed currency/flow/dimension combinations.
// An explicitly requested currency permits an empty ungrouped statistic; without
// that evidence an entirely empty ledger must not invent a currency.
func (a *assistantAggregation) fillEmpty() error {
	if len(a.combinations) == 0 && a.e.Args.Filter.Currency != "" && !containsString(a.e.Args.GroupBy, "account") && !containsString(a.e.Args.GroupBy, "category") {
		types := a.e.Args.Filter.Types
		if len(types) == 0 {
			types = []string{"income", "expense", "transfer", "balance_modification"}
		}
		for _, kind := range types {
			flows := []string{kind}
			if kind == "transfer" {
				flows = nil
				if a.e.Args.Filter.TransferSide != "destination" {
					flows = append(flows, "transfer_out")
				}
				if a.e.Args.Filter.TransferSide != "source" {
					flows = append(flows, "transfer_in")
				}
			}
			for _, flow := range flows {
				a.combinations[assistantAggregateDimension{Currency: a.e.Args.Filter.Currency, Flow: flow}] = true
			}
		}
	}
	flows := map[assistantAggregateDimension]bool{}
	for d := range a.combinations {
		flows[assistantAggregateDimension{Currency: d.Currency, Flow: d.Flow}] = true
	}
	for d := range flows {
		for p := range a.ranges {
			if err := a.e.Context.Err(); err != nil {
				return err
			}
			total := &assistantAggregateGroup{Period: p, Dimension: d}
			if key := a.storageKey(total); a.totals[key] == nil {
				a.totals[key] = total
			}
		}
	}
	return nil
}
