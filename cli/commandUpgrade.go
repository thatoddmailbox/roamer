package main

import (
	"fmt"
	"os"

	"github.com/thatoddmailbox/roamer"
)

func commandUpgrade(environment *roamer.Environment, options commandOptions, args []string) {
	requireSafe(environment)

	allMigrations, err := environment.ListAllMigrations()
	if err != nil {
		panic(err)
	}

	lastMigration, err := environment.GetLastAppliedMigration()
	if err != nil {
		panic(err)
	}

	if len(allMigrations) == 0 {
		fmt.Println("There are no migrations.")
		fmt.Println("Get started by doing `roamer create <description>`")
		os.Exit(1)
		return
	}

	latestMigration := allMigrations[len(allMigrations)-1]

	if lastMigration != nil {
		if latestMigration.ID == lastMigration.ID {
			fmt.Println("The database is already up-to-date.")
			return
		}
	}

	// we rewrite this as a go command to the latest migration
	commandGo(environment, options, []string{latestMigration.ID})
}
