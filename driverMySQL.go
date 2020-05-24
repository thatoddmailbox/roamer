package roamer

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
)

type driverMySQL struct {
	db     *sql.DB
	config *mysql.Config
}

func (d *driverMySQL) TableExists(name string) (bool, error) {
	rows, err := d.db.Query(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = ? AND table_name = ?",
		d.config.DBName,
		name,
	)
	defer rows.Close()
	if err != nil {
		return false, err
	}

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
