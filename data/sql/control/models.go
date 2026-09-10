package control

import "time"

func (u *SiteIdentity) LastSeenAtDate() *time.Time {
	if u.LastSeenAt == nil {
		return nil
	}
	t := time.UnixMilli(*u.LastSeenAt)
	return &t
}

func (u *SiteIdentity) LastWriteAtDate() *time.Time {
	if u.LastWriteAt == nil {
		return nil
	}
	t := time.UnixMilli(*u.LastWriteAt)
	return &t
}

func (u *Site) LastSeenAtDate() *time.Time {
	if u.LastSeenAt == nil {
		return nil
	}
	t := time.UnixMilli(*u.LastSeenAt)
	return &t
}

func (u *Site) LastWriteAtDate() *time.Time {
	if u.LastWriteAt == nil {
		return nil
	}
	t := time.UnixMilli(*u.LastWriteAt)
	return &t
}
