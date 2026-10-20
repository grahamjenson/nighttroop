package nsquery

import (
	"encoding/json"
	"strings"
)

type QueryStringArray []string

func (qsa *QueryStringArray) Stringp() *string {
	if qsa == nil {
		return nil
	}
	// Marshalled, not joined, so a quote or backslash in a value is escaped
	b, err := json.Marshal([]string(*qsa))
	if err != nil {
		return nil
	}
	res := string(b)
	return &res
}

func (qsa *QueryStringArray) ToArray() []string {
	if qsa == nil {
		return []string{}
	}
	return []string(*qsa)
}
func (qsa *QueryStringArray) UnmarshalJSON(data []byte) error {
	// First try a JSON array.
	var arr []string
	if err := json.Unmarshal(data, &arr); err == nil {
		*qsa = arr
		return nil
	}

	// Otherwise expect a JSON string.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	var parts []string
	switch {
	case s == "_all":
		parts = []string{}
	case strings.Contains(s, "|"):
		parts = strings.Split(s, "|")
	case strings.Contains(s, ","):
		parts = strings.Split(s, ",")
	default:
		parts = []string{s}
	}

	// Trim whitespace and drop empty elements.
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}

	*qsa = out
	return nil
}

type QuerySort string

var (
	// dates
	SortOccurredAtAsc  QuerySort = "occurred_at"
	SortOccurredAtDesc QuerySort = "-occurred_at"
	SortCreatedAtAsc   QuerySort = "created_at"
	SortCreatedAtDesc  QuerySort = "-created_at"
	SortUpdatedAtAsc   QuerySort = "updated_at"
	SortUpdatedAtDesc  QuerySort = "-updated_at"

	// Model Specific
	NameAsc  QuerySort = "name"
	NameDesc QuerySort = "-name"
)

func (qs *QuerySort) ToIntVal() int64 {
	if qs == nil {
		return -2
	}
	switch *qs {
	case SortOccurredAtAsc:
		return 1
	case SortOccurredAtDesc:
		return -1
	case SortCreatedAtAsc:
		return 2
	case SortCreatedAtDesc:
		return -2
	case SortUpdatedAtAsc:
		return 3
	case SortUpdatedAtDesc:
		return -3
	case NameAsc:
		return 4
	case NameDesc:
		return -4
	default:
		return -2
	}
}

var allowedSortFields = []QuerySort{
	SortOccurredAtAsc,
	SortOccurredAtDesc,
	SortCreatedAtAsc,
	SortCreatedAtDesc,
	SortUpdatedAtAsc,
	SortUpdatedAtDesc,
	NameAsc,
	NameDesc,
}

func (qs QuerySort) String() string {
	return string(qs)
}

func (qs *QuerySort) Support(allowed ...QuerySort) QuerySort {
	if len(allowed) == 0 && qs == nil {
		return SortCreatedAtDesc
	}
	if len(allowed) == 0 && qs != nil {
		return *qs
	}

	if qs == nil {
		return allowed[0]
	}

	for _, a := range allowed {
		if *qs == a {
			return *qs
		}
	}
	// We are returning first element as default
	return allowed[0]
}
