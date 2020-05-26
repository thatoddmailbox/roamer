package roamer

import (
	"database/sql"
	"errors"
	"io/ioutil"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ErrMigrationNotFound is returned when an attempt is made to get a migration that does not exist.
var ErrMigrationNotFound = errors.New("roamer: could not find the requested migration")

var reMigrationDescription = regexp.MustCompile("-- Description: (.*)\r*\n")

// A Migration represents a distinct operation performed on a database.
type Migration struct {
	ID          string
	Description string

	Index int

	downPath string
	upPath   string
}

// An AppliedMigration represents a history entry, describing a migration that had been applied to the database.
type AppliedMigration struct {
	ID        string
	AppliedAt int
	Dirty     bool
}

// ApplyMigration applies the migration to the database.
func (e *Environment) ApplyMigration(migration Migration, up bool) error {
	hasHistoryTable, err := e.driver.TableExists(tableNameRoamerHistory)
	if err != nil {
		return err
	}

	if !hasHistoryTable {
		// create the history table first
		_, err := e.db.Exec("CREATE TABLE " + tableNameRoamerHistory + `(
			id VARCHAR(20) PRIMARY KEY,
			appliedAt INT(11),
			dirty TINYINT(1)
			)`)
		if err != nil {
			return err
		}
	}

	if up {
		_, err = e.db.Exec(
			"INSERT INTO "+tableNameRoamerHistory+"(id, appliedAt, dirty) VALUES(?, ?, 1)",
			migration.ID,
			time.Now().Unix(),
		)
		if err != nil {
			return err
		}
	} else {
		_, err = e.db.Exec(
			"UPDATE "+tableNameRoamerHistory+" SET dirty = 1 WHERE id = ?",
			migration.ID,
		)
		if err != nil {
			return err
		}
	}

	// now read the migration file
	fileToRead := migration.downPath
	if up {
		fileToRead = migration.upPath
	}
	migrationData, err := e.readFile(fileToRead)
	if err != nil {
		return err
	}

	_, err = e.db.Exec(string(migrationData))
	if err != nil {
		return err
	}

	if up {
		_, err = e.db.Exec(
			"UPDATE "+tableNameRoamerHistory+" SET dirty = 0 WHERE id = ?",
			migration.ID,
		)
		if err != nil {
			return err
		}
	} else {
		_, err = e.db.Exec(
			"DELETE FROM "+tableNameRoamerHistory+" WHERE id = ?",
			migration.ID,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateMigration creates a new migration with the given name.
func (e *Environment) CreateMigration(description string) error {
	if e.pathOnDisk == "" {
		return errors.New("roamer: cannot create migration with this http.FileSystem")
	}

	id := strconv.FormatInt(time.Now().Unix(), 10)
	normalizedName := strings.Replace(strings.ToLower(description), " ", "_", -1)

	downPath := path.Join(e.pathOnDisk, id+"_"+normalizedName+"_down.sql")
	upPath := path.Join(e.pathOnDisk, id+"_"+normalizedName+"_up.sql")

	contents := "-- Description: " + description + "\n-- "

	err := ioutil.WriteFile(downPath, []byte(contents+"Down migration\n\n"), 0777)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(upPath, []byte(contents+"Up migration\n\n"), 0777)
	if err != nil {
		return err
	}

	return nil
}

// GetMigrationByID gets the migration with the given ID.
func (e *Environment) GetMigrationByID(id string) (Migration, error) {
	migration, ok := e.migrationsByID[id]
	if !ok {
		return Migration{}, ErrMigrationNotFound
	}

	return migration, nil
}

// ListAllMigrations gets all of the migrations defined in the migrations directory.
func (e *Environment) ListAllMigrations() ([]Migration, error) {
	return e.migrations, nil
}

// ListAppliedMigrations gets all of the migrations that have been applied to the database.
func (e *Environment) ListAppliedMigrations() ([]AppliedMigration, error) {
	tableExists, err := e.driver.TableExists(tableNameRoamerHistory)
	if err != nil {
		return nil, err
	}
	if !tableExists {
		return []AppliedMigration{}, nil
	}

	result := []AppliedMigration{}

	rows, err := e.db.Query("SELECT id, appliedAt, dirty FROM " + tableNameRoamerHistory + " ORDER BY id ASC")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		appliedMigration := AppliedMigration{}
		err = rows.Scan(&appliedMigration.ID, &appliedMigration.AppliedAt, &appliedMigration.Dirty)
		if err != nil {
			return nil, err
		}
		result = append(result, appliedMigration)
	}

	return result, nil
}

// GetLastAppliedMigration gets the last migration that has been applied to the database, returning nil if nothing has been applied.
func (e *Environment) GetLastAppliedMigration() (*AppliedMigration, error) {
	tableExists, err := e.driver.TableExists(tableNameRoamerHistory)
	if err != nil {
		return nil, err
	}
	if !tableExists {
		return nil, nil
	}

	result := AppliedMigration{}

	err = e.db.QueryRow(
		"SELECT id, appliedAt, dirty FROM "+tableNameRoamerHistory+" ORDER BY appliedAt DESC, id DESC LIMIT 1",
	).Scan(&result.ID, &result.AppliedAt, &result.Dirty)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &result, nil
}

// BeginTransaction begins a new transaction, which can then be used to apply migrations.
func (e *Environment) BeginTransaction() (*sql.Tx, error) {
	return e.db.Begin()
}
