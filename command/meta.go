package command

import (
	"flag"
	"github.com/mitchellh/cli"
)

type FlagSetFlags uint

const (
	FlagSetNone    FlagSetFlags = 0
	FlagSetClient  FlagSetFlags = 1 << iota
	FlagSetDefault              = FlagSetClient
)

type Meta struct {
	Ui cli.Ui
}

func (m *Meta) FlagSet(n string, fs FlagSetFlags) *flag.FlagSet {
	flags := flag.NewFlagSet(n, flag.ExitOnError)
	return flags
}
