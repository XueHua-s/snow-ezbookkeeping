package services

import (
	"bytes"
	"encoding/json"
	"github.com/mayswind/ezbookkeeping/pkg/llm/data"
	"io"
	"sort"
	"strconv"
	"unicode/utf8"
)

// assistantDomain is the single definition used for both schemas and strict
// runtime shape validation. Cross-field financial constraints live below.
type assistantDomain struct {
	Kind        string
	Fields      map[string]*assistantDomain
	Item        *assistantDomain
	Required    bool
	Enum        []string
	Description string
	MinItems    int
	Minimum     *int64
	Maximum     *int64
}

func domainString(values ...string) *assistantDomain {
	return &assistantDomain{Kind: "string", Enum: values}
}

func domainArray(item *assistantDomain) *assistantDomain {
	return &assistantDomain{Kind: "array", Item: item}
}

func domainObject(fields map[string]*assistantDomain) *assistantDomain {
	return &assistantDomain{Kind: "object", Fields: fields}
}

func requiredDomain(d *assistantDomain) *assistantDomain { d.Required = true; return d }

func assistantToolDomain(name data.ToolName) *assistantDomain {
	idArray := func() *assistantDomain {
		d := domainArray(requiredDomain(domainString()))
		d.MinItems = 1
		d.Description = "Positive integer IDs encoded as strings; null means no restriction, an empty array is invalid. Parent selections expand to effective descendants."
		return d
	}
	filter := domainObject(map[string]*assistantDomain{
		"start_time": domainString(), "end_time": domainString(), "types": domainArray(requiredDomain(domainString("income", "expense", "transfer", "balance_modification"))),
		"account_ids": idArray(), "category_ids": idArray(), "tag_ids": idArray(), "tag_match": domainString("any", "all"), "no_tags": {Kind: "boolean"},
		"currency": domainString(), "transfer_side": domainString("source", "destination", "either"), "amount_min": domainString(), "amount_max": domainString(), "keyword": domainString(),
	})
	filter.Description = "All effective historical and future transactions by default, including hidden accounts. Time bounds are second-precision RFC3339, start inclusive and end exclusive. Never infer a time cutoff."
	filter.Fields["types"].MinItems = 1
	filter.Fields["tag_match"].Description = "any (default) or all; requires nonempty tag_ids. no_tags=true conflicts with tag_ids."
	filter.Fields["transfer_side"].Description = "source, destination, or either (default). Account, currency and amount must match the same side; a transfer remains one logical transaction."
	filter.Fields["amount_min"].Description = "Inclusive signed decimal amount with at most two fractional digits; requires currency."
	filter.Fields["amount_max"].Description = filter.Fields["amount_min"].Description
	filter.Fields["keyword"].Description = "Case-insensitive literal comment substring; %, _ and backslash are literal characters."
	minimum, maximum := int64(1), int64(200)
	fields := map[string]*assistantDomain{"cursor": domainString(), "page_size": {Kind: "integer", Minimum: &minimum, Maximum: &maximum, Description: "Default 50; affects returned page only, never full-set totals."}}
	fields["cursor"].Description = "Opaque authenticated cursor from this tool; expires in 24 hours. Other parameters may be omitted; explicit parameters must match the original query. Use only for a requested continuation."
	order := func(values ...string) *assistantDomain {
		return domainObject(map[string]*assistantDomain{"field": requiredDomain(domainString(values...)), "direction": requiredDomain(domainString("asc", "desc"))})
	}
	switch name {
	case data.ToolListAccounts, data.ToolListTags:
		fields["keyword"] = domainString()
	case data.ToolListCategories:
		fields["keyword"] = domainString()
		fields["type"] = domainString("expense", "income", "transfer")
	case data.ToolGetAccountBalances:
		fields["account_ids"] = idArray()
	case data.ToolGetTransaction:
		return domainObject(map[string]*assistantDomain{"transaction_id": requiredDomain(domainString())})
	case data.ToolQueryTransactions:
		fields["filter"] = filter
		fields["order_by"] = order("time", "amount")
	case data.ToolAggregateTransactions:
		fields["filter"] = filter
		fields["order_by"] = order("group_key", "count", "sum", "min", "max")
		fields["group_by"] = domainArray(requiredDomain(domainString("account", "category", "day", "month", "year")))
		fields["metrics"] = domainArray(requiredDomain(domainString("count", "sum", "min", "max")))
		fields["metrics"].MinItems = 1
		fields["metrics"].Description = "Default count and sum; calculated across the complete matching ledger, separately by currency and flow."
		fields["periods"] = domainArray(requiredDomain(domainObject(map[string]*assistantDomain{"name": requiredDomain(domainString()), "start_time": requiredDomain(domainString()), "end_time": requiredDomain(domainString())})))
		fields["periods"].MinItems = 1
		fields["periods"].Description = "Named nonempty half-open periods with unique names; replaces filter.start_time/end_time."
		fields["comparison"] = domainObject(map[string]*assistantDomain{"baseline": requiredDomain(domainString()), "current": requiredDomain(domainString())})
	default:
		return nil
	}
	return domainObject(fields)
}

func (d *assistantDomain) schema() map[string]any {
	out := map[string]any{}
	if d.Description != "" {
		out["description"] = d.Description
	}
	if d.MinItems > 0 {
		out["minItems"] = d.MinItems
	}
	if d.Minimum != nil {
		out["minimum"] = *d.Minimum
	}
	if d.Maximum != nil {
		out["maximum"] = *d.Maximum
	}
	if d.Required {
		out["type"] = d.Kind
	} else {
		out["type"] = []string{d.Kind, "null"}
	}
	if len(d.Enum) > 0 {
		v := make([]any, 0, len(d.Enum)+1)
		for _, s := range d.Enum {
			v = append(v, s)
		}
		if !d.Required {
			v = append(v, nil)
		}
		out["enum"] = v
	}
	if d.Kind == "object" {
		props := map[string]any{}
		keys := make([]string, 0, len(d.Fields))
		for k, v := range d.Fields {
			props[k] = v.schema()
			keys = append(keys, k)
		}
		sort.Strings(keys)
		out["properties"] = props
		out["required"] = keys
		out["additionalProperties"] = false
	}
	if d.Item != nil {
		out["items"] = d.Item.schema()
	}
	return out
}

func (d *assistantDomain) validate(v any, path string) error {
	if v == nil {
		if d.Required {
			return assistantInvalid(path)
		}
		return nil
	}
	switch d.Kind {
	case "object":
		m, ok := v.(map[string]any)
		if !ok {
			return assistantInvalid(path)
		}
		for k, x := range m {
			child, ok := d.Fields[k]
			if !ok {
				return assistantInvalid(path + "." + k)
			}
			if err := child.validate(x, path+"."+k); err != nil {
				return err
			}
		}
		for k, child := range d.Fields {
			if child.Required && m[k] == nil {
				return assistantInvalid(path + "." + k)
			}
		}
	case "array":
		a, ok := v.([]any)
		if !ok {
			return assistantInvalid(path)
		}
		if len(a) < d.MinItems {
			return assistantInvalid(path)
		}
		for _, x := range a {
			if err := d.Item.validate(x, path); err != nil {
				return err
			}
		}
	case "string":
		s, ok := v.(string)
		if !ok {
			return assistantInvalid(path)
		}
		if len(d.Enum) > 0 && !containsString(d.Enum, s) {
			return assistantInvalid(path)
		}
	case "boolean":
		if _, ok := v.(bool); !ok {
			return assistantInvalid(path)
		}
	case "integer":
		n, ok := v.(json.Number)
		if !ok {
			return assistantInvalid(path)
		}
		parsed, err := strconv.ParseInt(string(n), 10, 64)
		if err != nil || (d.Minimum != nil && parsed < *d.Minimum) || (d.Maximum != nil && parsed > *d.Maximum) {
			return assistantInvalid(path)
		}
	}
	return nil
}

func assistantInvalid(field string) error {
	return &data.ToolError{Code: data.ToolErrorInvalidArgument, Message: "Invalid tool argument", Field: field}
}

func assistantStrictJSON(raw []byte) (map[string]any, error) {
	if !utf8.Valid(raw) {
		return nil, assistantInvalid("arguments")
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var read func(int) (any, error)
	read = func(depth int) (any, error) {
		if depth > 32 {
			return nil, assistantInvalid("arguments")
		}
		t, err := dec.Token()
		if err != nil {
			return nil, assistantInvalid("arguments")
		}
		delim, ok := t.(json.Delim)
		if !ok {
			return t, nil
		}
		switch delim {
		case '{':
			m := map[string]any{}
			for dec.More() {
				k, e := dec.Token()
				if e != nil {
					return nil, assistantInvalid("arguments")
				}
				key, ok := k.(string)
				if !ok {
					return nil, assistantInvalid("arguments")
				}
				if _, exists := m[key]; exists {
					return nil, assistantInvalid("duplicate_key")
				}
				v, e := read(depth + 1)
				if e != nil {
					return nil, e
				}
				m[key] = v
			}
			_, err = dec.Token()
			return m, err
		case '[':
			a := []any{}
			for dec.More() {
				v, e := read(depth + 1)
				if e != nil {
					return nil, e
				}
				a = append(a, v)
			}
			_, err = dec.Token()
			return a, err
		}
		return nil, assistantInvalid("arguments")
	}
	v, err := read(0)
	if err != nil {
		return nil, err
	}
	if _, err = dec.Token(); err != io.EOF {
		return nil, assistantInvalid("arguments")
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, assistantInvalid("arguments")
	}
	return m, nil
}
