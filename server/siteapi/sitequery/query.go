package sitequery

import (
	"net/url"

	"github.com/mitchellh/mapstructure"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/server/nightscoutapi/nsquery"
	"nighttroop.t1d.nz/utils/types"
)

//go:generate goverter gen .

// MaxQueryLimit is the most models one query returns, so no request can read a whole table at once
const MaxQueryLimit = 1000

type Query struct {
	Limit int64 `json:"limit,omitempty"`
	Skip  int64 `json:"skip,omitempty"`
}

// QueryValues are the query string values, before they are checked and mapped to a Query
type QueryValues struct {
	Limit *types.FlexibleInt64 `mapstructure:"limit,omitempty" json:"limit,omitempty"`
	Skip  *types.FlexibleInt64 `mapstructure:"skip,omitempty" json:"skip,omitempty"`
}

// goverter:converter
// goverter:output:format function
// goverter:output:file ./z_converter.go
type QueryConverter interface {
	// goverter:map . Limit | GetLimitInt64
	// goverter:map . Skip | GetSkipInt64
	QueryValuesToQuery(source QueryValues) Query

	// goverter:useZeroValueOnPointerInconsistency
	// goverter:ignore IncludeDeleted
	QueryToAllGlucoseReadings(source *Query) sitedb.AllGlucoseReadingsParams
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

func ParseQueryValues(values url.Values) (QueryValues, error) {
	params := map[string]string{}
	for key := range values {
		params[key] = values.Get(key)
	}

	var queryValues QueryValues
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &queryValues,
		DecodeHook: nsquery.JSONUnmarshalerHook,
	})
	if err != nil {
		return QueryValues{}, err
	}
	err = decoder.Decode(params)
	return queryValues, err
}
