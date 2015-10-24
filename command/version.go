package command

import (
	"bytes"
	"fmt"
	"github.com/mitchellh/cli"
)

type VersionCommand struct {
	Revision         string
	Version          string
	VersionPrerelase string
	Ui               cli.Ui
}

func (c *VersionCommand) Help() string {
	return ""
}

func (c *VersionCommand) Run(_ []string) int {
	var versionString bytes.Buffer
	fmt.Fprintf(&versionString, "Gypsy v%s", c.Version)
	if c.VersionPrerelase != "" {
		fmt.Fprintf(&versionString, "-%s", c.VersionPrerelase)
		if c.Revision != "" {
			fmt.Fprintf(&versionString, " (%s)", c.Revision)
		}
	}
	c.Ui.Output(versionString.String())
	return 0
}

func (c *VersionCommand) Synopsis() string {
	return "Prints Gypsy version"
}
