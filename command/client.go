package command

import (
	"github.com/ranjib/gypsy/client"
	log "github.com/sirupsen/logrus"
)

type ClientCommand struct {
	Meta
}

func (c *ClientCommand) Help() string {
	return "gypsy client -pipeline PIPELINE_NAME"
}

func (c *ClientCommand) Synopsis() string {
	return "Runs gypsy client"
}

func (c *ClientCommand) Run(args []string) int {
	var pipelineName string
	flags := c.Meta.FlagSet("client", FlagSetClient)
	flags.Usage = func() { c.Ui.Output(c.Help()) }
	flags.StringVar(&pipelineName, "pipeline", "", "Name of target pipeline")
	if err := flags.Parse(args); err != nil {
		log.Errorf("Failes to parse flags: %v", err)
		return 1
	}
	return client.BuildPipeline(pipelineName)
}
