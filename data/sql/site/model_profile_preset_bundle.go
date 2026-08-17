package site

type ProfilePresetBundle []*ProfilePreset

func (c ProfilePresetBundle) GetID() string {
	if len(c) == 0 || c[0] == nil {
		return ""
	}
	return c[0].ProfileStoreID
}
func (c ProfilePresetBundle) GetDeletedAt() *int64 {
	return nil

}
func (c ProfilePresetBundle) GetCreatedAt() int64 {
	if len(c) == 0 || c[0] == nil {
		return 0
	}
	return c[0].CreatedAt
}
func (c ProfilePresetBundle) GetUpdatedAt() int64 {
	if len(c) == 0 || c[0] == nil {
		return 0
	}
	return c[0].UpdatedAt
}

func (c ProfilePresetBundle) GetCreatedByIdentityName() string {
	if len(c) == 0 || c[0] == nil {
		return ""
	}
	return c[0].CreatedByIdentityName
}

func (c ProfilePresetBundle) GetUtcOffsetMinutes() *int64 {
	if len(c) == 0 || c[0] == nil {
		return nil
	}
	return c[0].UtcOffsetMinutes
}
func (c ProfilePresetBundle) GetOccurredAt() int64 {
	if len(c) == 0 || c[0] == nil {
		return 0
	}
	return c[0].OccurredAt
}

func (c ProfilePresetBundle) GetCreatedByAppName() *string {
	if len(c) == 0 || c[0] == nil {
		return nil
	}
	return c[0].CreatedByAppName
}

func (c ProfilePresetBundle) GetCreatedByDeviceName() *string {
	if len(c) == 0 || c[0] == nil {
		return nil
	}
	return c[0].CreatedByDeviceName
}

func (c ProfilePresetBundle) GetRawJsonData() *string {
	if len(c) == 0 || c[0] == nil {
		return nil
	}
	return c[0].RawJsonData
}

///////
// Many to Many
///////

func ManyPresetsToMayWrappers(source []*ProfilePreset) []*ProfilePresetBundle {
	if source == nil {
		return nil
	}
	lookup := map[string][]*ProfilePreset{}
	for _, preset := range source {
		if preset == nil {
			continue
		}
		if _, ok := lookup[preset.ProfileStoreID]; !ok {
			lookup[preset.ProfileStoreID] = []*ProfilePreset{}
		}
		if preset.IsDefault {
			// If this is the default profile, we want it to be first in the list
			lookup[preset.ProfileStoreID] = append([]*ProfilePreset{preset}, lookup[preset.ProfileStoreID]...)

		} else {
			lookup[preset.ProfileStoreID] = append(lookup[preset.ProfileStoreID], preset)
		}
	}

	wrappers := []*ProfilePresetBundle{}
	for _, presets := range lookup {
		wrapper := ProfilePresetBundle(presets)
		wrappers = append(wrappers, &wrapper)
	}
	return wrappers
}
