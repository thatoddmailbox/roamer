package roamer

import (
	"errors"
	"fmt"
)

// An Operation describes a series of migrations, bringing the database up or down to a new state.
type Operation struct {
	From *Migration
	To   *Migration

	Direction
	Distance int

	Stamp bool

	PreMigrationCallback func(*Migration, Direction)

	hasRun bool

	e         *Environment
	fromIndex int
	toIndex   int
}

// An OperationError is returned when there's an error while applying a certain migration.
type OperationError struct {
	Migration *Migration
	Inner     error
}

// Error returns a string representation of the OperationError.
func (e OperationError) Error() string {
	return fmt.Sprintf("roamer: migration %s: %s", e.Migration.ID, e.Inner.Error())
}

// NewOperation creates a new operation, with the given endpoints, in the environment.
func (e *Environment) NewOperation(from *Migration, to *Migration) (*Operation, error) {
	o := Operation{
		From: from,
		To:   to,

		Stamp: false,

		e: e,
	}

	if from != nil {
		_, exists := e.migrationsByID[from.ID]
		if !exists {
			return nil, ErrMigrationNotFound
		}
	}
	if to != nil {
		_, exists := e.migrationsByID[to.ID]
		if !exists {
			return nil, ErrMigrationNotFound
		}
	}

	// figure out the index of the from migration and the to migration
	o.fromIndex = -1
	o.toIndex = -1
	for i, migration := range e.migrations {
		if from != nil && migration.ID == from.ID {
			o.fromIndex = i
		}

		if to != nil && migration.ID == to.ID {
			o.toIndex = i
		}
	}

	o.Direction = DirectionDown
	if o.toIndex > o.fromIndex {
		o.Direction = DirectionUp
	}

	o.Distance = o.toIndex - o.fromIndex
	if o.Distance < 0 {
		o.Distance = -1 * o.Distance
	}

	return &o, nil
}

// DistanceString returns a string repesentation of the distance spanned by the Operation.
func (o *Operation) DistanceString() string {
	plural := ""
	if o.Distance != 1 {
		plural = "s"
	}

	return fmt.Sprintf(
		"%d %s migration%s",
		o.Distance,
		o.Direction.String(),
		plural,
	)
}

// Run runs the given operation.
func (o *Operation) Run() error {
	if o.hasRun {
		return errors.New("roamer: operation has already been run")
	}

	lastApplied, err := o.e.GetLastAppliedMigration()
	if err != nil {
		return err
	}
	if lastApplied == nil {
		if o.From != nil {
			return errors.New("roamer: cannot run operation with incorrect From migration")
		}
	}
	if lastApplied != nil {
		if o.From == nil || o.From.ID != lastApplied.ID {
			return errors.New("roamer: cannot run operation with incorrect From migration")
		}
	}

	o.hasRun = true

	offset := 0
	if o.Direction == DirectionUp {
		offset = 1
	}

	for i := o.fromIndex; i != o.toIndex; i += int(o.Direction) {
		migrationToApply := o.e.migrations[i+offset]

		if o.PreMigrationCallback != nil {
			o.PreMigrationCallback(&migrationToApply, o.Direction)
		}

		err := o.e.ApplyMigration(migrationToApply, o.Direction, o.Stamp)
		if err != nil {
			// the migration failed!
			return OperationError{
				Migration: &migrationToApply,
				Inner:     err,
			}
		}
	}

	return nil
}
