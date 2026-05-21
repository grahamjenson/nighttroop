package control

import "time"

func (u *User) LastSeenAtDate() *time.Time {
	if u.LastSeenAt == nil {
		return nil
	}
	t := time.UnixMilli(*u.LastSeenAt)
	return &t
}
