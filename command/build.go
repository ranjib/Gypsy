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

// Encapsulates all gypsy command line interfaces
package command

import (
	"github.com/ranjib/gypsy/build"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type BuildCommand struct {
	Meta
}

func (c *BuildCommand) Help() string {
	helpText := `
   Usage: gypsy build PIPELINE_NAME -run_id RUN_ID

	 General Options:
	` + generalOptionsUage()
	return strings.TrimSpace(helpText)
}

func (c *BuildCommand) Synopsis() string {
	return "Runs gypsy build agent"
}

func (c *BuildCommand) Run(args []string) int {
	var pipelineName string
	var runId int
	flags := c.Meta.FlagSet("client", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.StringVar(&pipelineName, "pipeline", "", "Name of target pipeline")
	flags.IntVar(&runId, "run_id", 0, "Run ID")
	if err := flags.Parse(args); err != nil {
		log.Errorf("Failes to parse flags: %v", err)
		return 1
	}
	var logOutput io.Writer
	if c.Meta.logOutput != "" {
		fi, err := os.OpenFile(c.Meta.logOutput, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Errorf("Failed to open log output file '%s'. Error: %s\n", c.Meta.logOutput, err)
			return -1
		}
		defer fi.Close()
		logOutput = fi
	} else {
		logOutput = os.Stdout
	}
	util.ConfigureLogging(c.Meta.logLevel, c.Meta.logFormat, logOutput)
	if pipelineName == "" {
		log.Errorf("Must provide a valid pipeline name")
		return 1
	}
	if runId == 0 {
		log.Errorf("Must provide a valid run id")
		return 1
	}
	return build.BuildPipeline(pipelineName, runId)
}
