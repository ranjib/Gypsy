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
	"fmt"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type ListPipelineCommand struct {
	Meta
}

func (c *ListPipelineCommand) Help() string {
	helpString := `
	Usage: gypsy list-pipeline [gypsy.yml]"

	General Options:
	` + generalOptionsUsage()
	return strings.TrimSpace(helpString)
}

func (c *ListPipelineCommand) Synopsis() string {
	return "List present pipelines"
}

func (c *ListPipelineCommand) Run(args []string) int {
	flags := c.Meta.FlagSet("list-pipeline", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	if err := flags.Parse(args); err != nil {
		log.Errorf("Failed to parse cli arguments. Error: %s\n", err)
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
	client, err := c.Meta.Client()
	if err != nil {
		log.Errorf("Failed to create api client. Error:%s\n", err)
		return -1
	}
	pipelines, err := client.ListPipelines()
	if err != nil {
		log.Errorf("Failed to list pipeline. Error: %s\n", err)
		return -1
	}
	c.Ui.Output(fmt.Sprintf("%6s %-30s", "#", "pipeline"))
	for i, pipeline := range pipelines {
		c.Ui.Output(fmt.Sprintf("%6d %-30s", i+1, pipeline))
	}
	return 0
}
