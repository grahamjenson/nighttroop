package apiquery

//go:generate goverter gen .

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converter.go
type QueryConverter interface {
	// goverter:map . Limit | GetLimitInt64
	// goverter:map . Skip | GetSkipInt64
	// goverter:map . Hours | GetHoursInt64
	// goverter:map . Type | GetType
	QueryValuesToQuery(source QueryValues) Query
}

// GetLimitInt64 defaults a missing or bad limit to MaxQueryLimit, and caps it there
func GetLimitInt64(source QueryValues) int64 {
	if source.Limit == nil || *source.Limit <= 0 || *source.Limit > MaxQueryLimit {
		return MaxQueryLimit
	}
	return source.Limit.Int64()
}

func GetSkipInt64(source QueryValues) int64 {
	if source.Skip == nil || *source.Skip < 0 {
		return 0
	}
	return source.Skip.Int64()
}

// GetHoursInt64 leaves a missing or bad hours unset, and caps it at MaxQueryHours
func GetHoursInt64(source QueryValues) *int64 {
	if source.Hours == nil || *source.Hours <= 0 {
		return nil
	}
	hours := min(source.Hours.Int64(), MaxQueryHours)
	return &hours
}

// GetType leaves an empty type unset, which is every type
func GetType(source QueryValues) *string {
	if source.Type == nil || *source.Type == "" {
		return nil
	}
	return source.Type
}
