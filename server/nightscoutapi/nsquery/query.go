package nsquery

// Because we cant support all mongo queries and filters
// Nighttroop is just going to enumerate and translate all queries into a limited set of supported queries

type Query struct {
	// Dates
	// lte and gte both
	DateGTE *int64 `json:"date_gte,omitempty"`
	DateGT  *int64 `json:"date_gt,omitempty"`
	DateEQ  *int64 `json:"date_eq,omitempty"`
	DateLTE *int64 `json:"date_lte,omitempty"`
	DateLT  *int64 `json:"date_lt,omitempty"`

	CreatedAtGTE *int64 `json:"created_at_gte,omitempty"`
	CreatedAtGT  *int64 `json:"created_at_gt,omitempty"`
	CreatedAtLTE *int64 `json:"created_at_lte,omitempty"`
	CreatedAtLT  *int64 `json:"created_at_lt,omitempty"`

	UpdatedGTE *int64 `json:"updated_gte,omitempty"`
	UpdatedGT  *int64 `json:"updated_gt,omitempty"`
	UpdatedLTE *int64 `json:"updated_lte,omitempty"`
	UpdatedLT  *int64 `json:"updated_lt,omitempty"`

	// Type
	TypeEq  *string `json:"type_eq,omitempty"`
	TypeNeq *string `json:"type_neq,omitempty"`
	TypeIn  *string `json:"type_in,omitempty"`
	TypeNin *string `json:"type_nin,omitempty"`

	// Identity
	IDEq *string `json:"id_eq,omitempty"`
	IDIn *string `json:"id_in,omitempty"`

	DeviceEq *string `json:"device_eq,omitempty"`
	AppEq    *string `json:"app_eq,omitempty"`

	IncludeDeleted bool `json:"includeDeleted,omitempty"`

	Raw bool `json:"raw,omitempty"`
	// Altering Return values or structure
	// Limits
	Limit int64 `json:"limit,omitempty"`
	Skip  int64 `json:"skip,omitempty"`
	// CountGiven is whether a limit or count was sent, as Limit is the most a search returns without one
	CountGiven bool `json:"count_given,omitempty"`

	/////
	// MODEL SPECIFIC QUERIES
	/////

	// Entries
	SgvEQ  *int64 `json:"sgv_eq,omitempty"`
	SgvGTE *int64 `json:"sgv_gte,omitempty"`
	SgvGT  *int64 `json:"sgv_gt,omitempty"`
	SgvLTE *int64 `json:"sgv_lte,omitempty"`
	SgvLT  *int64 `json:"sgv_lt,omitempty"`

	// Treatments
	CarbsEQ  *int64 `json:"carbs_eq,omitempty"`
	CarbsGTE *int64 `json:"carbs_gte,omitempty"`
	CarbsGT  *int64 `json:"carbs_gt,omitempty"`
	CarbsLTE *int64 `json:"carbs_lte,omitempty"`
	CarbsLT  *int64 `json:"carbs_lt,omitempty"`

	InsulinEQ  *int64 `json:"insulin_eq,omitempty"`
	InsulinGTE *int64 `json:"insulin_gte,omitempty"`
	InsulinGT  *int64 `json:"insulin_gt,omitempty"`
	InsulinLTE *int64 `json:"insulin_lte,omitempty"`
	InsulinLT  *int64 `json:"insulin_lt,omitempty"`

	DurationEQ  *int64 `json:"duration_eq,omitempty"`
	DurationGTE *int64 `json:"duration_gte,omitempty"`
	DurationGT  *int64 `json:"duration_gt,omitempty"`
	DurationLTE *int64 `json:"duration_lte,omitempty"`
	DurationLT  *int64 `json:"duration_lt,omitempty"`

	NameEq *string `json:"name_eq,omitempty"`

	CarbsExists    *bool   `json:"carbs_exists,omitempty"`
	DurationExists *bool   `json:"duration_exists,omitempty"`
	EnteredByEq    *string `json:"entered_by_eq,omitempty"`
	EnteredByNin   *string `json:"entered_by_nin,omitempty"`

	// Util Options
	HasFilters bool       `json:"has_filters,omitempty"`
	SortOrder  *QuerySort `json:"sort_order,omitempty"`
	Fields     []string   `json:"fields,omitempty"`
	Envelope   *bool      `json:"envelope,omitempty"`

	// UnknownFilter is whether a filter is on a field or with an operator Nighttroop does not know, so nothing matches
	UnknownFilter bool `json:"unknown_filter,omitempty"`
}
