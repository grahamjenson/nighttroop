package nsquery

// Because we cant support all mongo queries and filters
// Nighttroop is just going to enumerate and translate all queries into a limited set of supported queries

type Query struct {
	// Dates
	// lte and gte both
	DateGTE *int64
	DateGT  *int64
	DateEQ  *int64
	DateLTE *int64
	DateLT  *int64

	CreatedAtGTE *int64
	CreatedAtGT  *int64
	CreatedAtLTE *int64
	CreatedAtLT  *int64

	UpdatedGTE *int64
	UpdatedGT  *int64
	UpdatedLTE *int64
	UpdatedLT  *int64

	// Type
	TypeEq  *string
	TypeNeq *string
	TypeIn  *string
	TypeNin *string

	// Identity
	IDEq *string
	IDIn *string

	DeviceEq *string
	AppEq    *string

	IncludeDeleted bool

	Raw bool
	// Altering Return values or structure
	// Limits
	Limit int64
	Skip  int64

	/////
	// MODEL SPECIFIC QUERIES
	/////

	// Entries
	SgvEQ  *int64
	SgvGTE *int64
	SgvGT  *int64
	SgvLTE *int64
	SgvLT  *int64

	// Treatments
	CarbsEQ  *int64
	CarbsGTE *int64
	CarbsGT  *int64
	CarbsLTE *int64
	CarbsLT  *int64

	InsulinEQ  *int64
	InsulinGTE *int64
	InsulinGT  *int64
	InsulinLTE *int64
	InsulinLT  *int64

	DurationEQ  *int64
	DurationGTE *int64
	DurationGT  *int64
	DurationLTE *int64
	DurationLT  *int64

	NameEq *string

	// Util Options
	HasFilters bool
	SortOrder  *QuerySort
	Fields     []string
	Envelope   *bool
}
