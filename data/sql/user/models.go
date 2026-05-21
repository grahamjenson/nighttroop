package user

import "time"

func (model *Entry) CreatedAtDate() time.Time {
	t := time.UnixMilli(model.CreatedAt)
	return t
}
