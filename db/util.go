package db

import (
	"database/sql"
	"log"
)

func CheckRowsAffected(r sql.Result, expectedRows int64) error {
	n, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		log.Println("CKC ", n)
		return ErrUnexpectedRowCount
	}
	return nil
}
