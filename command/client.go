package command

type ClientCommand struct {
	Meta
}

func (c *ClientCommand) Help() string {
	return ""
}

func (c *ClientCommand) Synopsis() string {
	return ""
}

func (c *ClientCommand) Run(_ []string) int {
	return 0
}
