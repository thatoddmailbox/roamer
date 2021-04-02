package main

import (
	"fmt"
	"os"
	"path"

	"github.com/thatoddmailbox/roamer"
)

func commandSetup(environment *roamer.Environment, options commandOptions, args []string) {
	// find the default configs
	localConfig := roamer.DefaultLocalConfig

	// build paths
	basePath := args[0]
	localConfigFile := "roamer." + args[1] + ".toml"
	localConfigPath := path.Join(basePath, localConfigFile)

	// check that nothing is there
	_, err := os.Stat(localConfigPath)
	if err == nil {
		fmt.Println("A " + localConfigFile + " file already exists!")
		fmt.Println("This normally means you already have a roamer environment set up.")
		os.Exit(1)
	}
	if !os.IsNotExist(err) {
		panic(err)
	}

	// now actually create the thing
	err = writeTOMLToFile(localConfigPath, 0600, localConfig)
	if err != nil {
		panic(err)
	}

	fmt.Println("A " + localConfigFile + " file has been created for you.")
	fmt.Println("You should edit it to include your database connection details.")
}
