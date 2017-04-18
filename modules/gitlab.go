package modules

import (
	"fmt"

	"github.com/urfave/cli"
)

//Gitlab struct todo
type Gitlab struct {
}

//LoadFlags for cli
func (*Gitlab) LoadFlags() []cli.Command {

	var commands []cli.Command = make([]cli.Command, 0)
	n := cli.Command{
		Name:    "Gitlab",
		Aliases: []string{"g"},
		Usage:   "options for task templates",
		Subcommands: []cli.Command{
			{
				Name:  "version",
				Usage: "Get the gitlab module version",
				Action: func(c *cli.Context) error {
					fmt.Println("0.0.1")
					// fmt.Println("new task template: ", c.Args().First())
					return nil
				},
			},
		},
	}

	commands = append(commands, n)
	return commands
}
