package roamer

// A Migration represents a distinct operation performed on a database.
type Migration struct {
	ID          string
	Description string

	downFile string
	upFile   string
}

func (e *Environment) ListAllMigrations() ([]Migration, error) {
	return e.migrations, nil
}

func (e *Environment) ListAppliedMigrations() ([]Migration, error) {
	return []Migration{}, nil
}
