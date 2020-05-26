package main

import (
	"fmt"
	"os"

	"github.com/AlecAivazis/survey"

	"github.com/thatoddmailbox/roamer"
)

func commandGo(environment *roamer.Environment, options commandOptions, args []string) {
	requireSafe(environment)

	lastAppliedMigration, err := environment.GetLastAppliedMigration()
	if err != nil {
		panic(err)
	}

	var lastMigration *roamer.Migration
	if lastAppliedMigration != nil {
		lastMigration, err = environment.ResolveIDOrOffset(lastAppliedMigration.ID)
		if err != nil {
			if err == roamer.ErrMigrationNotFound {
				fmt.Printf("Last applied migration %s does not exist.\nDo `roamer status` for help resolving this.\n", lastAppliedMigration.ID)
				os.Exit(1)
				return
			} else {
				panic(err)
			}
		}
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

	operation, err := environment.NewOperation(lastMigration, targetMigration)
	if err != nil {
		panic(err)
	}

	fromString := "[nothing]"
	if lastMigration != nil {
		fromString = lastMigration.ID
	}
	toString := "[nothing]"
	if targetMigration != nil {
		toString = targetMigration.ID
	}
	fmt.Printf("Going %s -> %s (%s)\n\n", fromString, toString, operation.DistanceString())

	if operation.Direction == roamer.DirectionDown {
		if !options.force {
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

	operation.PreMigrationCallback = func(m *roamer.Migration, d roamer.Direction) {
		fmt.Printf("Applying %s migration %s - %s\n", d.String(), m.ID, m.Description)
	}

	err = operation.Run()
	if err != nil {
		operationErr, isOperationErr := err.(roamer.OperationError)
		if isOperationErr {
			fmt.Printf("There was an error applying migration %s!\n", operationErr.Migration.ID)
			fmt.Println()
			fmt.Println(operationErr.Inner)
			fmt.Println()
			fmt.Println("The database may now be in an inconsistent state. The migration has been marked as dirty.")
			fmt.Println("You must connect to the database and manually resolve the issue.")
			fmt.Println("Then, update the " + environment.GetHistoryTableName() + " table and, depending on how you resolved the issue, either delete the migration or set the dirty flag to 0.")
			os.Exit(1)
		}

		panic(err)
	}

	fmt.Printf("\nThe database is now at migration %s.\n", toString)
}
