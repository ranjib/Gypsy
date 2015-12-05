package main

import (
	"github.com/mitchellh/cli"
	"github.com/ranjib/gypsy/command"
	"os"
)

func Commands() map[string]cli.CommandFactory {
	meta := command.Meta{
		Ui: &cli.BasicUi{
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}
	return map[string]cli.CommandFactory{
		"client": func() (cli.Command, error) {
			return &command.ClientCommand{
				Meta: meta,
			}, nil
		},
		"server": func() (cli.Command, error) {
			return &command.ServerCommand{
				Meta: meta,
			}, nil
		},
		"dockerfile": func() (cli.Command, error) {
			return &command.DockerfileCommand{
				Meta: meta,
			}, nil
		},
		"version": func() (cli.Command, error) {
			return &command.VersionCommand{
				Revision:         GitCommit,
				Version:          Version,
				VersionPrerelase: VersionPrerelase,
				Ui:               meta.Ui,
			}, nil
		},
	}
}
