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

	orderMatches, err := environment.VerifyOrder()
	if err != nil {
		panic(err)
	}
	if !orderMatches {
		fmt.Println("The migrations on disk do not match the order of migrations applied to the database.")
		fmt.Println("The status command is currently unable to provide useful output in this scenario.")
		fmt.Println("You should check the " + environment.GetHistoryTableName() + " table and compare it to the migrations on disk.")
		os.Exit(1)
	}

	maxIDLen := 0
	for _, migration := range appliedMigrations {
		if len(migration.ID) > maxIDLen {
			maxIDLen = len(migration.ID)
		}
	}
	for _, migration := range allMigrations {
		if len(migration.ID) > maxIDLen {
			maxIDLen = len(migration.ID)
		}
	}

	columnSpacingStr := "    "

	idColumnPadding := strings.Repeat(" ", maxIDLen-2+1)

	fmt.Println("ID" + idColumnPadding + columnSpacingStr + "Description")

	haveDirty := false
	haveMissing := false

	i := 0
	for _, appliedMigration := range appliedMigrations {
		idDisplay := appliedMigration.ID + " "
		if appliedMigration.Dirty {
			haveDirty = true
			idDisplay = "!" + appliedMigration.ID
		}

		migration, err := environment.GetMigrationByID(appliedMigration.ID)
		if err == nil {
			fmt.Println(idDisplay + columnSpacingStr + migration.Description)
		} else {
			if err == roamer.ErrMigrationNotFound {
				fmt.Println(idDisplay + columnSpacingStr + "*** ERROR: missing corresponding migration file!")
				haveMissing = true
			} else {
				panic(err)
			}
		}

		i += 1
	}

	unappliedMigrations := allMigrations
	if i <= len(allMigrations) {
		unappliedMigrations = allMigrations[i:]
	}

	for _, unappliedMigration := range unappliedMigrations {
		fmt.Println("*" + unappliedMigration.ID + columnSpacingStr + unappliedMigration.Description)
	}

	if len(allMigrations) != len(appliedMigrations) {
		fmt.Println()
		fmt.Println("(* = migration has not been applied)")
	}

	if haveDirty {
		fmt.Println()
		fmt.Println("(! = migration is dirty)")
		fmt.Println("One or more migrations are marked as dirty. The database may be in an inconsistent state.")
		fmt.Println("You must connect to the database and manually resolve the issue.")
		fmt.Println("Then, update the " + environment.GetHistoryTableName() + " table and, depending on how you resolved the issue, either delete the migration or set the dirty flag to 0.")
	}

	if haveMissing {
		fmt.Println()
		fmt.Println("One or more applied migrations do not have a matching file on disk. Are you using the correct environment?")
		fmt.Println("You should restore these files, or, if you know what you're doing, remove the migration entries from the " + environment.GetHistoryTableName() + " table.")
	}

	if haveDirty || haveMissing {
		os.Exit(1)
	}
}
