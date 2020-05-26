package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/AlecAivazis/survey"

	"github.com/thatoddmailbox/roamer"
)

func commandGo(environment *roamer.Environment, force bool, args []string) {
	requireSafe(environment)

	allMigrations, err := environment.ListAllMigrations()
	if err != nil {
		panic(err)
	}

	lastMigration, err := environment.GetLastAppliedMigration()
	if err != nil {
		panic(err)
	}

	targetMigration, err := environment.ResolveIDOrOffset(args[0])
	if err != nil {
		if err == roamer.ErrMigrationNotFound {
			fmt.Printf("Migration %s does not exist.\n", args[0])
			os.Exit(1)
			return
		} else {
			panic(err)
		}
	}

	if lastMigration != nil && targetMigration != nil {
		if lastMigration.ID == targetMigration.ID {
			fmt.Printf("The database is already at migration %s.\n", targetMigration.ID)
			os.Exit(1)
			return
		}
	}

	if lastMigration == nil && targetMigration == nil {
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

		if targetMigration != nil && migration.ID == targetMigration.ID {
			targetMigrationIndex = i
		}
	}

	if targetMigration == nil {
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
	if targetMigration != nil {
		toString = targetMigration.ID
	}
	fmt.Printf("Going %s -> %s (%s)\n\n", fromString, toString, distanceString)

	if !directionIsUp {
		if !force {
			answer := false
			survey.AskOne(&survey.Confirm{
				Message: "You're about to run one or more down migrations, which can result in data loss. Continue?",
			}, &answer)

			fmt.Println()

			if !answer {
				fmt.Println("Migration cancelled. No changes have been made.")
				os.Exit(1)
				return
			}
		}
	}

	offset := 0
	if directionIsUp {
		offset = 1
	}

	for i := lastAppliedMigrationIndex; i != targetMigrationIndex; i += direction {
		migrationToApply := allMigrations[i+offset]
		fmt.Printf("Applying %s migration %s - %s\n", directionString, migrationToApply.ID, migrationToApply.Description)

		err = environment.ApplyMigration(migrationToApply, directionIsUp)
		if err != nil {
			// the migration failed!
			fmt.Printf("There was an error applying migration %s!\n", migrationToApply.ID)
			fmt.Println()
			fmt.Println(err)
			fmt.Println()
			fmt.Println("The database may now be in an inconsistent state. The migration has been marked as dirty.")
			fmt.Println("You must connect to the database and manually resolve the issue.")
			fmt.Println("Then, update the " + environment.GetHistoryTableName() + " table and, depending on how you resolved the issue, either delete the migration or set the dirty flag to 0.")
			os.Exit(1)
		}
	}

	fmt.Printf("\nThe database is now at migration %s.\n", toString)
}
