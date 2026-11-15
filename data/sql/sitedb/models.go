package sitedb

func (model FoodPreset) GetDeletedAt() *int64 {
	return nil
}

// GetUtcOffsetMinutes is nil, as a food keeps no offset, so it is given the site's
func (model FoodPreset) GetUtcOffsetMinutes() *int64 {
	return nil
}

func (model FoodPreset) GetOccurredAt() int64 {
	return model.CreatedAt
}

func (model ProfilePreset) GetDeletedAt() *int64 {
	return nil
}
