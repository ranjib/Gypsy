package command

import (
	"github.com/ranjib/gypsy/util"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

type DeletePipelineCommand struct {
	Meta
}

func (c *DeletePipelineCommand) Help() string {
	helpString := `
	Usage: gypsy delete-pipeline <pipeline name>"

	General Options:
	` + generalOptionsUage()
	return strings.TrimSpace(helpString)
}

func (c *DeletePipelineCommand) Synopsis() string {
	return "Delete a pipeline"
}

func (c *DeletePipelineCommand) Run(args []string) int {
	var file string
	flags := c.Meta.FlagSet("create-pipeline", FlagSetClient)
	flags.StringVar(&file, "file", "gypsy.yml", "Path to yaml file containing pipeline details")
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
	args = flags.Args()
	if len(args) != 1 {
		c.Ui.Error(c.Help())
		return -1
	}
	client, err := c.Meta.Client()
	if err != nil {
		log.Errorf("Failed to create api client. Error:%s\n", err)
		return -1
	}
	if err := client.DeletePipeline(args[0]); err != nil {
		log.Errorf("Failed to delete pipeline. Error:%s\n", err)
		return -1
	}
	c.Ui.Output("Suscessfully deleted " + args[0])
	return 0
}
