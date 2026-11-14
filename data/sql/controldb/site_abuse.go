package controldb

// AbuseSizeFactor is how many times its size limit a site's DB can be before the site is flagged for abuse. The server
// truncates a site near its limit, so a DB far over it is likely bloated on purpose, to hurt the server and its backups.
const AbuseSizeFactor = 3

// FlaggedForAbuse is whether the site may be abused, so the server limits it harder. Each check of abuse is added here,
// so all that limit an abused site use the same flag.
func (s *Site) FlaggedForAbuse() bool {
	return s.overAbuseSize()
}

// overAbuseSize is whether the site's DB is over AbuseSizeFactor times its size limit, false if either is not known
func (s *Site) overAbuseSize() bool {
	return s.EstimatedDbSizeKb != nil && s.MaxDbSizeKbLimit != nil && *s.EstimatedDbSizeKb > AbuseSizeFactor*(*s.MaxDbSizeKbLimit)
}
