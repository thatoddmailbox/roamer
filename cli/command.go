package main

import (
	"github.com/thatoddmailbox/roamer"
)

type commandAction func(*roamer.Environment, []string)

type command struct {
	Name        string
	Description string
	Arguments   []string
	Action      commandAction
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
