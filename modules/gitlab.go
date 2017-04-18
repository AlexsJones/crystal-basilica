package modules

import "github.com/urfave/cli"

type Gitlab struct {
}

func (*Gitlab) LoadFlags() []cli.Command {

  var commands []cli.Command = make([]cli.Command, 0)
  n := cli.Command{
      Name:        "doo",
      Aliases:     []string{"do"},
      Category:    "motion",
      Usage:       "do the doo",
      UsageText:   "doo - does the dooing",
      Description: "no really, there is a lot of dooing to be done",
      ArgsUsage:   "[arrgh]",
      Flags: []cli.Flag{
        cli.BoolFlag{Name: "forever, forevvarr"},
      },
    }
  commands = append(commands, n)
  return commands
}
