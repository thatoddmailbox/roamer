package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/thatoddmailbox/roamer"
)

func printHelp() {
	fmt.Printf("Usage: %s <command>\n\n", os.Args[0])
	fmt.Println("Flags:")
	flag.PrintDefaults()
	fmt.Println("")
	fmt.Println("Commands:")
	for _, commandName := range commandOrder {
		command := commands[commandName]
		fmt.Printf("  %s\n", command.Name)
		fmt.Printf("        %s\n", command.Description)
	}

	os.Exit(0)
}

func main() {
	flagHelp := flag.Bool("help", false, "Display usage information.")
	flagVersion := flag.Bool("version", false, "Display the current version.")
	flagEnvironment := flag.String("env", "./", "The directory to use as an environment.")
	flagForce := flag.Bool("force", false, "Skip any prompts for down migrations. Useful for shell scripts that run migrate.")
	flagLocalConfig := flag.String("local-config", "local", "The file to use as the local config.")
	flag.Parse()

	registerCommands()

	if *flagVersion {
		fmt.Printf("roamer version %s\n", roamer.GetVersionString())
		os.Exit(0)
		return
	}

	args := flag.Args()

	if len(args) == 0 || *flagHelp {
		printHelp()
		return
	}

	envInfo, err := os.Stat(*flagEnvironment)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Environment '%s' does not exist.\n", *flagEnvironment)
			os.Exit(1)
			return
		}

		panic(err)
	}
	if !envInfo.IsDir() {
		fmt.Printf("Environment '%s' is actually a file!\n", *flagEnvironment)
		fmt.Println("Make sure your environment is the directory containing your roamer.toml file, not the file itself!")
		os.Exit(1)
		return
	}

	command, commandExists := commands[args[0]]
	if !commandExists {
		fmt.Printf("Unknown command '%s'. Do -help to see all commands.\n", args[0])
		os.Exit(1)
		return
	}

	// verify argument count
	if len(args)-1 != len(command.Arguments) {
		fmt.Printf("Incorrect usage of '%s'. Do -help to see usage information.\n", args[0])
		os.Exit(1)
		return
	}

	var environment *roamer.Environment

	// init and setup are special cases, don't load the environment for it
	if command.Name != "init" && command.Name != "setup" {
		environment, err = roamer.NewEnvironmentFromDisk(*flagEnvironment, *flagLocalConfig)
		if err != nil {
			panic(err)
		}
	} else {
		// sneak in the environment path as an argument
		// a bit of a hack but it works
		args = []string{command.Name, *flagEnvironment}
	}

	command.Action(environment, commandOptions{*flagForce}, args[1:])
}
