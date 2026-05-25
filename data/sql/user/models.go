package user

import "time"

func (model *Sgv) CreatedAtDate() time.Time {
	t := time.UnixMilli(model.CreatedAt)
	return t
}
