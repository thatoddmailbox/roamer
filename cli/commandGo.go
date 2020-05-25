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

	if targetMigrationID != "none" {
		_, err = environment.GetMigrationByID(targetMigrationID)
		if err != nil {
			if err == roamer.ErrMigrationNotFound {
				fmt.Printf("Migration %s does not exist.\n", targetMigrationID)
				os.Exit(1)
				return
			} else {
				panic(err)
			}
		}
	}

	if lastMigration != nil && targetMigrationID != "none" {
		if lastMigration.ID == targetMigrationID {
			fmt.Printf("The database is already at migration %s.\n", targetMigrationID)
			os.Exit(1)
			return
		}
	}

	if lastMigration == nil && targetMigrationID == "none" {
		fmt.Println("The database is already at no migrations.")
		os.Exit(1)
		return
	}

	// figure out the index of the current and the target
	lastAppliedMigrationIndex := -1
	targetMigrationIndex := 0
	for i, migration := range allMigrations {
		if lastMigration != nil && migration.ID == lastMigration.ID {
			lastAppliedMigrationIndex = i
		}

		if targetMigrationID != "none" && migration.ID == targetMigrationID {
			targetMigrationIndex = i
		}
	}

	if targetMigrationID == "none" {
		targetMigrationIndex = -1
	}

	// determine the direction
	direction := -1
	directionIsUp := false
	directionString := "down"
	if targetMigrationIndex > lastAppliedMigrationIndex {
		direction = 1
		directionIsUp = true
		directionString = "up"
	}

	distance := targetMigrationIndex - lastAppliedMigrationIndex
	if distance < 0 {
		distance = -1 * distance
	}

	distanceString := ""
	distanceString += strconv.Itoa(distance) + " "
	distanceString += directionString
	distanceString += " migration"
	if distance != 1 {
		distanceString += "s"
	}

	fromString := "[nothing]"
	if lastMigration != nil {
		fromString = lastMigration.ID
	}
	toString := "[nothing]"
	if targetMigrationID != "none" {
		toString = targetMigrationID
	}
	fmt.Printf("Going %s -> %s (%s)\n\n", fromString, toString, distanceString)

	tx, err := environment.BeginTransaction()
	if err != nil {
		panic(err)
	}

	offset := 0
	if directionIsUp {
		offset = 1
	}

	for i := lastAppliedMigrationIndex; i != targetMigrationIndex; i += direction {
		migrationToApply := allMigrations[i+offset]
		fmt.Printf("Applying %s migration %s - %s\n", directionString, migrationToApply.ID, migrationToApply.Description)

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
