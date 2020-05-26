package main

import (
	"fmt"
	"os"

	"github.com/thatoddmailbox/roamer"
)

type commandAction func(*roamer.Environment, commandOptions, []string)

type command struct {
	Name        string
	Description string
	Arguments   []string
	Action      commandAction
}

type commandOptions struct {
	force bool
}

var commands map[string]command
var commandOrder []string

func registerCommand(cmd command) {
	commands[cmd.Name] = cmd
	commandOrder = append(commandOrder, cmd.Name)
}

func registerCommands() {
	commands = map[string]command{}
	commandOrder = []string{}

	registerCommand(command{
		Name:        "create",
		Description: "Create a new migration",
		Arguments:   []string{"NAME"},
		Action:      commandCreate,
	})
	registerCommand(command{
		Name:        "go",
		Description: "Migrates the database to the given migration",
		Arguments:   []string{"MIGRATION ID"},
		Action:      commandGo,
	})
	registerCommand(command{
		Name:        "init",
		Description: "Sets up a new environment",
		Arguments:   []string{},
		Action:      commandInit,
	})
	registerCommand(command{
		Name:        "setup",
		Description: "Sets up an existing environment with database configuration options",
		Arguments:   []string{},
		Action:      commandSetup,
	})
	registerCommand(command{
		Name:        "status",
		Description: "Gets the currently applied migration in the database",
		Arguments:   []string{},
		Action:      commandStatus,
	})
	registerCommand(command{
		Name:        "upgrade",
		Description: "Upgrade the database to the latest version",
		Arguments:   []string{},
		Action:      commandUpgrade,
	})
}

func requireSafe(environment *roamer.Environment) {
	isClean, err := environment.VerifyNoDirty()
	if err != nil {
		panic(err)
	}

	if !isClean {
		fmt.Println("One or more migrations are marked as dirty.")
		fmt.Println("It is not safe to apply additional migrations at this time.")
		fmt.Println("For more information, and help resolving the issue, do `roamer status`.")
		os.Exit(1)
	}

	allExist, err := environment.VerifyExist()
	if err != nil {
		panic(err)
	}

	if !allExist {
		fmt.Println("There are migrations in the database that do not exist on disk.")
		fmt.Println("It is not safe to apply additional migrations at this time.")
		fmt.Println("For more information, and help resolving the issue, do `roamer status`.")
		os.Exit(1)
	}

	isInOrder, err := environment.VerifyOrder()
	if err != nil {
		panic(err)
	}

	if !isInOrder {
		fmt.Println("The migrations on disk do not match the order of migrations applied to the database.")
		fmt.Println("It is not safe to apply additional migrations at this time.")
		fmt.Println("For more information, and help resolving the issue, do `roamer status`.")
		os.Exit(1)
	}
}
