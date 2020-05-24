package main

import (
	"fmt"

	"github.com/thatoddmailbox/roamer"
)

func commandCreate(environment *roamer.Environment, args []string) {
	description := args[0]
	err := environment.CreateMigration(description)
	if err != nil {
		panic(err)
	}

	fmt.Println("A new migration file has been created.")
}
