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
	"github.com/ranjib/gypsy/structs"
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

type CreatePipelineCommand struct {
	Meta
}

func (c *CreatePipelineCommand) Help() string {
	helpString := `
	Usage: gypsy create-pipeline [gypsy.yml]"

	General Options:
	` + generalOptionsUsage()
	return strings.TrimSpace(helpString)
}

func (c *CreatePipelineCommand) Synopsis() string {
	return "Create a pipeline from yaml formatted file"
}

func (c *CreatePipelineCommand) Run(args []string) int {
	flags := c.Meta.FlagSet("create-pipeline", FlagSetClient)
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
	file := "gypsy.yml"
	args = flags.Args()
	if len(args) > 0 {
		file = args[0]
	}

	client, err := c.Meta.Client()
	if err != nil {
		log.Errorf("Failed to create api client. Error:%s\n", err)
		return -1
	}

	content, err := ioutil.ReadFile(file)
	if err != nil {
		log.Errorf("Failed to read file %s . Error: %s\n", file, err)
		return -1
	}
	var pipeline structs.Pipeline
	if err := yaml.Unmarshal(content, &pipeline); err != nil {
		log.Errorf("Failed to unmarshal pipeline : %v", err)
		return -1
	}

	if err := client.CreatePipeline(&pipeline); err != nil {
		log.Errorf("Failed to create pipeline. Error: %s\n", err)
		return -1
	}
	c.Ui.Output("Sucessfully created pipeline " + pipeline.Name)
	return 0
}
