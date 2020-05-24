package roamer

import (
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"
)

// A Migration represents a distinct operation performed on a database.
type Migration struct {
	ID          string
	Description string

	downFile string
	upFile   string
}

func (e *Environment) CreateMigration(description string) error {
	id := strconv.FormatInt(time.Now().Unix(), 10)
	normalizedName := strings.Replace(strings.ToLower(description), " ", "_", -1)

	downPath := path.Join(e.fullMigrationsPath, id+"_"+normalizedName+"_down.sql")
	upPath := path.Join(e.fullMigrationsPath, id+"_"+normalizedName+"_up.sql")

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

func (e *Environment) ListAllMigrations() ([]Migration, error) {
	return e.migrations, nil
}

func (e *Environment) ListAppliedMigrations() ([]Migration, error) {
	return []Migration{}, nil
}
