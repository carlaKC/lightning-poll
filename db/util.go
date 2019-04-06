package db

import (
	"database/sql"
)

func CheckRowsAffected(r sql.Result, expectedRows int64) error {
	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return ErrUnexpectedRowCount
	}
	return nil
}
