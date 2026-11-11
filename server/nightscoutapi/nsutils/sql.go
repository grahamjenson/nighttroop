package nsutils

type SQLResultCustomResult struct {
	LastInsertIdInt int64
	RowsAffectedInt int64
	Err             error
}

func (s SQLResultCustomResult) LastInsertId() (int64, error) {
	return s.LastInsertIdInt, s.Err
}

func (s SQLResultCustomResult) RowsAffected() (int64, error) {
	return s.RowsAffectedInt, s.Err
}
