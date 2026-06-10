package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"nighttroop.t1d.nz/utils/consts"
)

type SettingValue[T any] struct {
	Key     string
	Default T
}

var (
	TimezoneSetting = SettingValue[Timezone]{
		Key: "timezone",
		Default: Timezone{
			TZ: *time.UTC,
		},
	}

	UnitsSetting = SettingValue[Units]{
		Key: "units",
		Default: Units{
			Units: consts.UNITS_MG_DL,
		},
	}
)

type Timezone struct {
	TZ time.Location `json:"name"`
}

type Units struct {
	Units consts.Units `json:"units"`
}

func GetSetting[T any](
	ctx context.Context,
	q *Queries,
	setting SettingValue[T],
) (T, error) {
	var result T

	value, err := q.getSettingFromDB(ctx, setting.Key)

	if err != nil && err == sql.ErrNoRows {
		return setting.Default, nil
	}

	if err != nil {
		return result, err
	}

	err = json.Unmarshal(value, &result)
	return result, err
}

func SetSetting[T any](
	ctx context.Context,
	q *Queries,
	setting SettingValue[T],
	value T,
) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return q.upsertSettingToDB(ctx, setting.Key, string(data))
}
