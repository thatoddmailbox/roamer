package main

import (
	"fmt"
	"strings"

	"github.com/thatoddmailbox/roamer"
)

func commandStatus(environment *roamer.Environment, args []string) {
	allMigrations, err := environment.ListAllMigrations()
	if err != nil {
		panic(err)
	}

	appliedMigrations, err := environment.ListAppliedMigrations()
	if err != nil {
		panic(err)
	}

	if len(allMigrations) == 0 {
		fmt.Println("There are no migrations.")
		fmt.Println("Get started by doing `roamer create <description>`")
		return
	}

	maxIDLen := 0
	for _, migration := range allMigrations {
		if len(migration.ID) > maxIDLen {
			maxIDLen = len(migration.ID)
		}
	}

	columnSpacingStr := "    "

	idColumnPadding := strings.Repeat(" ", maxIDLen-2+1)

	fmt.Println("ID" + idColumnPadding + columnSpacingStr + "Description")

	i := 0
	for _, appliedMigration := range appliedMigrations {
		fmt.Println(appliedMigration.ID + columnSpacingStr + appliedMigration.Description)
		i += 1
	}

	for _, unappliedMigration := range allMigrations[i:] {
		fmt.Println("*" + unappliedMigration.ID + columnSpacingStr + unappliedMigration.Description)
	}
}
