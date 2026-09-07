package site

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
