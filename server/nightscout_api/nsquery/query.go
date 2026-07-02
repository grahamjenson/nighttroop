package nsquery

// Because we cant support all mongo queries and filters
// Nighttroop is just going to enumerate and translate all queries into a limited set of supported queries

type Query struct {
	// Dates
	// lte and gte both
	DateGTE *int64
	DateGT  *int64

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
	CarbsLT  *int64

	InsulinEQ  *int64
	InsulinGTE *int64
	InsulinLT  *int64

	NameEq         *string
	IncludeDeleted *bool

	// Altering Return values or structure
	// Limits
	Limit int64
	Skip  int64

	// Sort
	HasFilters bool
	SortOrder  *QuerySort
	Fields     []string
	Envelope   *bool
}
