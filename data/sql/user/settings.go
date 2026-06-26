package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"nighttroop.t1d.nz/utils/consts"
)

var settings = map[string]UserSetting{
	"timezone": TimezoneSetting,
	"units":    UnitsSetting,
}

func GetUserSettingFromKey(key string) (UserSetting, bool) {
	setting, ok := settings[key]
	return setting, ok
}

type UserSetting interface {
	GetKey() string
	GetValue(ctx context.Context, q *Queries) (any, error)
	GetValueModel(ctx context.Context, q *Queries) (*Setting, error)
	UpsertValue(ctx context.Context, q *Queries, value any) (*Setting, error)
	CreateValue(string) (any, error)
}

type timezoneSetting struct{}

var TimezoneSetting timezoneSetting

type TimezoneSettingValue struct {
	TZ time.Location `json:"name"`
}

func (s timezoneSetting) GetKey() string {
	return "timezone"
}

func (s timezoneSetting) UpsertValue(ctx context.Context, q *Queries, value any) (*Setting, error) {
	if value == nil {
		return nil, sql.ErrNoRows
	}
	_, ok := value.(TimezoneSettingValue)
	if !ok {
		return nil, fmt.Errorf("invalid Timezone value setting type %T", value)
	}
	return setSetting(ctx, q, s.GetKey(), value)
}

func (s timezoneSetting) GetValue(ctx context.Context, q *Queries) (any, error) {
	return getSettingValue(ctx, q, s.GetKey(), &TimezoneSettingValue{}, TimezoneSettingValue{TZ: *time.UTC})
}

func (s timezoneSetting) CreateValue(value string) (any, error) {
	x := &TimezoneSettingValue{}
	err := json.Unmarshal([]byte(value), x)
	if err != nil {
		return nil, err
	}
	return x, nil
}

func (s timezoneSetting) GetValueModel(ctx context.Context, q *Queries) (*Setting, error) {
	return q.getSettingFromDB(ctx, s.GetKey())
}

type unitsSetting struct{}

var UnitsSetting unitsSetting

type UnitsSettingValue struct {
	Units consts.Units `json:"units"`
}

func (s unitsSetting) GetKey() string {
	return "units"
}

func (s unitsSetting) UpsertValue(ctx context.Context, q *Queries, value any) (*Setting, error) {
	if value == nil {
		return nil, sql.ErrNoRows
	}
	_, ok := value.(UnitsSettingValue)
	if !ok {
		return nil, fmt.Errorf("invalid Unit value setting type %T", value)
	}
	return setSetting(ctx, q, s.GetKey(), value)
}

func (s unitsSetting) GetValue(ctx context.Context, q *Queries) (any, error) {
	return getSettingValue(ctx, q, s.GetKey(), &UnitsSettingValue{}, UnitsSettingValue{Units: consts.UNITS_MG_DL})
}

func (s unitsSetting) GetValueModel(ctx context.Context, q *Queries) (*Setting, error) {
	return q.getSettingFromDB(ctx, s.GetKey())
}

func (s unitsSetting) CreateValue(value string) (any, error) {
	x := &UnitsSettingValue{}
	err := json.Unmarshal([]byte(value), x)
	if err != nil {
		return nil, err
	}
	return x, nil
}

// ////
// Private Functions
// ////

func getSettingValue(
	ctx context.Context,
	q *Queries,
	key string,
	retValue any,
	defaultValue any,
) (any, error) {
	value, err := q.getSettingFromDB(ctx, key)

	if err != nil && err == sql.ErrNoRows {
		return defaultValue, nil
	}

	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(value.Value), retValue)
	return retValue, err
}

func setSetting(
	ctx context.Context,
	q *Queries,
	key string,
	value any,
) (*Setting, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return q.upsertSettingToDB(ctx, key, string(data))
}

// Neeed to fit some other interfaces
func (s *Setting) GetCreatedByIdentityName() string {
	return ""
}
