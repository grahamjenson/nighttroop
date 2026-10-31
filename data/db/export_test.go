package db

import "nighttroop.t1d.nz/utils/aslog"

// FlushFrom writes buffer's held logs, for a test to write its own buffer's rather than the server's shared one
func (l *LogsDB) FlushFrom(buffer *aslog.Buffer) { l.flush(buffer) }
