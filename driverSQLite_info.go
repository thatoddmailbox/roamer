// +build nocgo

package roamer

import (
	"database/sql"
	"errors"
)

const sqliteAvailable = false

type driverSQLite struct {
	db *sql.DB
}

func (d *driverSQLite) TableExists(name string) (bool, error) {
	return false, errors.New("roamer: sqlite support not available")
}
