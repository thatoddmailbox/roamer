package main

import (
	"fmt"
	"os"
	"path"

	"github.com/BurntSushi/toml"

	"github.com/thatoddmailbox/roamer"
)

func writeTOMLToFile(filePath string, perm os.FileMode, thing interface{}) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := toml.NewEncoder(file)
	encoder.Indent = ""
	return encoder.Encode(thing)
}

func commandInit(environment *roamer.Environment, options commandOptions, args []string) {
	// find the default configs
	config := roamer.DefaultConfig
	localConfig := roamer.DefaultLocalConfig

	// build paths
	basePath := args[0]
	configPath := path.Join(basePath, "roamer.toml")
	localConfigPath := path.Join(basePath, "roamer.local.toml")
	migrationsPath := path.Join(basePath, config.Environment.MigrationDirectory)

	// check that nothing is there
	_, err := os.Stat(configPath)
	if err == nil {
		fmt.Println("A roamer.toml file already exists!")
		fmt.Println("Perhaps you meant `roamer setup`?")
		os.Exit(1)
	}
	if !os.IsNotExist(err) {
		panic(err)
	}
	_, err = os.Stat(localConfigPath)
	if err == nil {
		fmt.Println("A roamer.local.toml file already exists!")
		fmt.Println("It looks like you already have a roamer environment set up.")
		os.Exit(1)
	}
	if !os.IsNotExist(err) {
		panic(err)
	}
	_, err = os.Stat(migrationsPath)
	if err == nil {
		fmt.Println("A migrations directory already exists!")
		fmt.Println("You need to remove or move this directory first.")
		os.Exit(1)
	}
	if !os.IsNotExist(err) {
		panic(err)
	}

	// now actually create the things
	err = writeTOMLToFile(configPath, 0775, config)
	if err != nil {
		panic(err)
	}
	err = writeTOMLToFile(localConfigPath, 0700, localConfig)
	if err != nil {
		panic(err)
	}
	err = os.Mkdir(migrationsPath, 0775)
	if err != nil {
		panic(err)
	}

	fmt.Println("A roamer.toml, roamer.local.toml, and migrations directory have been created for you.")
	fmt.Println("If you're using version control software, make sure to exclude roamer.local.toml from it!")
}
