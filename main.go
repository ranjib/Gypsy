package main

import (
	"github.com/mitchellh/cli"
	log "github.com/sirupsen/logrus"
	"os"
)

func main() {
	os.Exit(realMain(os.Args[1:], Commands()))
}

func realMain(args []string, commands map[string]cli.CommandFactory) int {
	log.Println("Everybody's here. Puke stinks like beer")
	for _, arg := range args {
		if arg == "-v" || arg == "-version" || arg == "--version" {
			newArgs := make([]string, len(args)+1)
			newArgs[0] = "version"
			copy(newArgs[1:], args)
			args = newArgs
			break
		}
	}
	cmdNames := make([]string, 0, len(commands))
	for cmdName, _ := range commands {
		cmdNames = append(cmdNames, cmdName)
	}
	cli := &cli.CLI{
		Args:     args,
		Commands: commands,
		HelpFunc: cli.FilteredHelpFunc(cmdNames, cli.BasicHelpFunc("gypsy")),
	}
	exitCode, err := cli.Run()
	if err != nil {
		log.Warnf("Error executing CLI: %s", err.Error())
		return 1
	}
	return exitCode
}
