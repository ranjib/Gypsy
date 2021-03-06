// Copyright 2015 Ranjib Dey.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package command

import (
	"flag"
	"github.com/mitchellh/cli"
	"github.com/ranjib/gypsy/api"
	"strings"
)

type FlagSetFlags uint

const (
	FlagSetNone   FlagSetFlags = 0
	FlagSetLog    FlagSetFlags = 1 << iota
	FlagSetClient              = FlagSetLog
)

type Meta struct {
	Ui        cli.Ui
	address   string
	logLevel  string
	logFormat string
	logOutput string
}

func (m *Meta) FlagSet(n string, fs FlagSetFlags) *flag.FlagSet {
	flags := flag.NewFlagSet(n, flag.ExitOnError)
	if fs&FlagSetLog != 0 {
		flags.StringVar(&m.logLevel, "loglevel", "info", "-loglevel <level>")
		flags.StringVar(&m.logFormat, "logformat", "text", "-logformat <format>")
		flags.StringVar(&m.logOutput, "logoutput", "", "-logoutput <file>")
	}
	if fs&FlagSetClient != 0 {
		flags.StringVar(&m.address, "address", "http://localhost:5678", "-address <gypsy server>")
	}
	return flags
}

func (m *Meta) Client() (*api.Client, error) {
	config := api.DefaultConfig()
	config.Address = m.address
	return api.NewClient(config)
}

func generalOptionsUsage() string {
	helpText := `
	-address=<addr>
		Address of gypsy server
		Default = http://localhost:5678
	-loglevel=<level>
		Set log level (can be debug, info, warn, error, fatal or panic)
		Default = info
	-logformat=<format>
		 Set log fromat (can be text or json)
		 Default = text
	-logoutput=<file>
		 Set log output file
		 Default = "" (logs to stdout)
	`
	return strings.TrimSpace(helpText)
}
