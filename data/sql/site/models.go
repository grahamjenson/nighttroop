package site

import (
	"strings"
	"time"
)

func (model *GlucoseReading) CreatedAtDate() time.Time {
	t := time.UnixMilli(model.CreatedAt)
	return t
}

func IsDuplicateID(err error) bool {
	// Exmaple Error String is "UNIQUE constraint failed: glucose_readings.id"
	// TODO: probably slow
	return strings.Contains(err.Error(), "UNIQUE constraint failed") && strings.Contains(err.Error(), ".id")
}

func (model FoodPreset) GetDeletedAt() *int64 {
	return nil
}

func (model FoodPreset) GetUtcOffsetMinutes() *int64 {
	utcOffset := int64(0)
	return &utcOffset
}

func (model FoodPreset) GetOccurredAt() int64 {
	return model.CreatedAt
}

func (model ProfilePreset) GetDeletedAt() *int64 {
	return nil
}
