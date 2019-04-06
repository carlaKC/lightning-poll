package db

import "github.com/pkg/errors"

var (
	ErrUnexpectedRowCount = errors.New("Unexpected number of rows updated")
	ErrNotFound           = errors.New("Not found.")
)
