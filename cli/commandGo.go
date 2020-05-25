package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/thatoddmailbox/roamer"
)

func commandGo(environment *roamer.Environment, args []string) {
	allMigrations, err := environment.ListAllMigrations()
	if err != nil {
		panic(err)
	}

	lastMigration, err := environment.GetLastAppliedMigration()
	if err != nil {
		panic(err)
	}

	targetMigrationID := args[0]

	_, err = environment.GetMigrationByID(targetMigrationID)
	if err != nil {
		if err == roamer.ErrMigrationNotFound {
			fmt.Printf("Migration %s does not exist.", targetMigrationID)
			os.Exit(1)
			return
		} else {
			panic(err)
		}
	}

	if lastMigration != nil {
		if lastMigration.ID == targetMigrationID {
			fmt.Printf("The database is already at migration %s.", targetMigrationID)
			os.Exit(1)
			return
		}
	}

	// figure out the index of the current and the target
	lastAppliedMigrationIndex := -1
	targetMigrationIndex := 0
	if lastMigration != nil {
		for i, migration := range allMigrations {
			if migration.ID == lastMigration.ID {
				lastAppliedMigrationIndex = i
			}

			if migration.ID == targetMigrationID {
				targetMigrationIndex = i
			}
		}
	}

	// determine the direction
	direction := -1
	directionIsUp := false
	if targetMigrationIndex > lastAppliedMigrationIndex {
		direction = 1
		directionIsUp = true
	}

	distance := targetMigrationIndex - lastAppliedMigrationIndex

	distanceString := ""
	distanceString += strconv.Itoa(distance) + " "
	if directionIsUp {
		distanceString += "up"
	} else {
		distanceString += "down"
	}
	distanceString += " migration"
	if distance != 1 {
		distanceString += "s"
	}

	fromString := "[nothing]"
	if lastMigration != nil {
		fromString = lastMigration.ID
	}
	fmt.Printf("Going %s -> %s (%s)\n\n", fromString, targetMigrationID, distanceString)

	tx, err := environment.BeginTransaction()
	if err != nil {
		panic(err)
	}

	for i := lastAppliedMigrationIndex; i != targetMigrationIndex; i += direction {
		migrationToApply := allMigrations[i+1]
		fmt.Printf("Applying migration %s - %s\n", migrationToApply.ID, migrationToApply.Description)

		err = environment.ApplyMigration(tx, migrationToApply, directionIsUp)
		if err != nil {
			tx.Rollback()
			panic(err)
		}
	}

	err = tx.Commit()
	if err != nil {
		panic(err)
	}

	fmt.Printf("\nThe database is now at migration %s.\n", targetMigrationID)
}
