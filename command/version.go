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
