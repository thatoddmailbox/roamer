// +build !nocgo

package roamer

import (
	"database/sql"
	"errors"

	_ "github.com/mattn/go-sqlite3"
)

const sqliteAvailable = true

type driverSQLite struct {
	db *sql.DB
}

func (d *driverSQLite) TableExists(name string) (bool, error) {
	rows, err := d.db.Query(
		"SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?",
		name,
	)
	if err != nil {
		return false, err
	}
	defer rows.Close()

	if !rows.Next() {
		return false, errors.New("roamer: did not expect no response to COUNT(*)")
	}

	count := 0
	err = rows.Scan(&count)
	if err != nil {
		return false, err
	}

	if count == 1 {
		return true, nil
	}

	return false, nil
}
