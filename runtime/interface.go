package runtime

import "github.com/urfave/cli"

type Module interface {
	LoadFlags() []cli.Command
}
