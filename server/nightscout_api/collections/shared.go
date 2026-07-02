package collections

import (
	"nighttroop.t1d.nz/server/nightscout_api/nsutils"
	"nighttroop.t1d.nz/utils"
)

func CreateRecordSlice[T any](in []T, fn func(T) nsutils.Record) []nsutils.Record {
	out := make([]nsutils.Record, len(in))
	for i, v := range in {
		out[i] = fn(v)
	}
	return out
}

func ptr[T any](v T) *T {
	return &v
}

type ConverterToRecordFN[T any, PT interface {
	*T
	GetID() string
	GetDeletedAt() *int64
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetCreatedByIdentityName() string
}] func(PT) nsutils.Record

func DBModelToRecord[T any, PT interface {
	*T
	GetID() string
	GetDeletedAt() *int64
	GetCreatedAt() int64
	GetUpdatedAt() int64
	GetCreatedByIdentityName() string
}, S any](collectionName nsutils.Collection, convertFunc func(PT) S) ConverterToRecordFN[T, PT] {
	return func(dbModel PT) nsutils.Record {
		if dbModel == nil {
			return nsutils.Record{}
		}

		nsModel := convertFunc(dbModel)

		mapModel, err := utils.ToMap(&nsModel)
		if err != nil {
			return nsutils.Record{}
		}

		record := nsutils.Record{
			Collection:  string(collectionName),
			ID:          dbModel.GetID(),
			Data:        mapModel,
			SrvCreated:  dbModel.GetCreatedAt(),
			SrvModified: dbModel.GetUpdatedAt(),
			Subject:     dbModel.GetCreatedByIdentityName(),
			IsValid:     true,
			DeletedAt:   dbModel.GetDeletedAt(),
		}
		return record
	}
}
