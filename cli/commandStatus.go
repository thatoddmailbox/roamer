package main

import (
	"fmt"
	"os"
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
		os.Exit(1)
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
		migration, err := environment.GetMigrationByID(appliedMigration.ID)
		if err == nil {
			fmt.Println(appliedMigration.ID + " " + columnSpacingStr + migration.Description)
		} else {
			if err == roamer.ErrMigrationNotFound {
				fmt.Println(appliedMigration.ID + columnSpacingStr + "*** ERROR: missing corresponding migration file!")
			} else {
				panic(err)
			}
		}

		i += 1
	}

	for _, unappliedMigration := range allMigrations[i:] {
		fmt.Println("*" + unappliedMigration.ID + columnSpacingStr + unappliedMigration.Description)
	}

	if len(allMigrations) != len(appliedMigrations) {
		fmt.Println()
		fmt.Println("(* = migration has not been applied)")
	}

}
