package modules

import (
	"fmt"

	"github.com/urfave/cli"
)

//Jenkins struct todo
type Jenkins struct {
}

//LoadFlags for cli
func (*Jenkins) LoadFlags() []cli.Command {

	var commands []cli.Command = make([]cli.Command, 0)
	n := cli.Command{
		Name:    "Jenkins",
		Aliases: []string{"j"},
		Usage:   "options for task templates",
		Subcommands: []cli.Command{
			{
				Name:  "version",
				Usage: "Get the jenkins module version",
				Action: func(c *cli.Context) error {
					fmt.Println("0.0.1")
					return nil
				},
			},
		},
	}

	commands = append(commands, n)
	return commands
}
