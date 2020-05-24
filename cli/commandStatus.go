package main

import (
	"log"

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

	log.Println(allMigrations)
	log.Println(appliedMigrations)
}
