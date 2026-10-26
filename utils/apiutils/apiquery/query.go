package apiquery

import (
	"net/url"

	"github.com/mitchellh/mapstructure"
	"nighttroop.t1d.nz/utils/types"
)

// MaxQueryLimit is the most models one query returns, so no request can read a whole table at once
const MaxQueryLimit = 1000

// MaxQueryHours is a week, the furthest back a summary goes
const MaxQueryHours = 7 * 24

type Query struct {
	Limit int64 `json:"limit,omitempty"`
	Skip  int64 `json:"skip,omitempty"`
	// Device, when set, keeps only what that device sent
	Device *string `json:"device,omitempty"`
	// Hours, when set, is how far back a summary goes, at most MaxQueryHours
	Hours *int64 `json:"hours,omitempty"`
	// Order, when set, is what a collection that can be ordered more than one way is ordered by
	Order *string `json:"order,omitempty"`
}

// QueryValues are the query string values, before they are checked and mapped to a Query
type QueryValues struct {
	Limit  *types.FlexibleInt64 `mapstructure:"limit,omitempty" json:"limit,omitempty"`
	Skip   *types.FlexibleInt64 `mapstructure:"skip,omitempty" json:"skip,omitempty"`
	Device *string              `mapstructure:"device,omitempty" json:"device,omitempty"`
	Hours  *types.FlexibleInt64 `mapstructure:"hours,omitempty" json:"hours,omitempty"`
	Order  *string              `mapstructure:"order,omitempty" json:"order,omitempty"`
}

func ParseQueryValues(values url.Values) (QueryValues, error) {
	params := map[string]string{}
	for key := range values {
		params[key] = values.Get(key)
	}

	var queryValues QueryValues
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:     &queryValues,
		DecodeHook: types.JSONUnmarshalerHook,
	})
	if err != nil {
		return QueryValues{}, err
	}
	err = decoder.Decode(params)
	return queryValues, err
}
