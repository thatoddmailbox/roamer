package main

import (
	"fmt"
	"os"
	"path"

	"github.com/thatoddmailbox/roamer"
)

func commandSetup(environment *roamer.Environment, force bool, args []string) {
	// find the default configs
	localConfig := roamer.DefaultLocalConfig

	// build paths
	basePath := args[0]
	localConfigPath := path.Join(basePath, "roamer.local.toml")

	// check that nothing is there
	_, err := os.Stat(localConfigPath)
	if err == nil {
		fmt.Println("A roamer.local.toml file already exists!")
		fmt.Println("It looks like you already have a roamer environment set up.")
		os.Exit(1)
	}
	if !os.IsNotExist(err) {
		panic(err)
	}

	// now actually create the thing
	err = writeTOMLToFile(localConfigPath, localConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println("A roamer.local.toml file has been created for you.")
	fmt.Println("You should edit it to include your database connection details.")
}
