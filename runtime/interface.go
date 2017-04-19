package runtime

import "github.com/urfave/cli"

//Module interface for common functions
type Module interface {
	LoadFlags() []cli.Command
}
