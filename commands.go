package main

import (
	"github.com/mitchellh/cli"
	"github.com/ranjib/gypsy/command"
	"os"
)

// Commands register all gypsy related commands
func Commands() map[string]cli.CommandFactory {
	meta := command.Meta{
		Ui: &cli.BasicUi{
			Writer:      os.Stdout,
			ErrorWriter: os.Stderr,
		},
	}
	return map[string]cli.CommandFactory{
		"build": func() (cli.Command, error) {
			return &command.BuildCommand{
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
		"create-pipeline": func() (cli.Command, error) {
			return &command.CreatePipelineCommand{
				Meta: meta,
			}, nil
		},
		"show-pipeline": func() (cli.Command, error) {
			return &command.ShowPipelineCommand{
				Meta: meta,
			}, nil
		},
		"delete-pipeline": func() (cli.Command, error) {
			return &command.DeletePipelineCommand{
				Meta: meta,
			}, nil
		},
		"list-pipelines": func() (cli.Command, error) {
			return &command.ListPipelineCommand{
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
