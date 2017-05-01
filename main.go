package main

import (
	"bytes"
	"os"

	modules "github.com/AlexsJones/schism/modules"
	runtime "github.com/AlexsJones/schism/runtime"
	"github.com/dimiro1/banner"
	"github.com/urfave/cli"
)

const b string = `
{{ .AnsiColor.Green }} ______   ______   ___   ___    ________  ______   ___ __ __
{{ .AnsiColor.Green }}/_____/\ /_____/\ /__/\ /__/\  /_______/\/_____/\ /__//_//_/\
{{ .AnsiColor.Green }}\::::_\/_\:::__\/ \::\ \\  \ \ \__.::._\/\::::_\/_\::\| \| \ \
{{ .AnsiColor.Green }} \:\/___/\\:\ \  __\::\/_\ .\ \   \::\ \  \:\/___/\\:.      \ \
{{ .AnsiColor.Green }}  \_::._\:\\:\ \/_/\\:: ___::\ \  _\::\ \__\_::._\:\\:.\-/\  \ \
{{ .AnsiColor.Green }}    /____\:\\:\_\ \ \\: \ \\::\ \/__\::\__/\ /____\:\\. \  \  \ \
{{ .AnsiColor.Green }}    \_____\/ \_____\/ \__\/ \::\/\________\/ \_____\/ \__\/ \__\/
{{ .AnsiColor.Default }}
`

func load(m runtime.Module, commands *[]cli.Command) {
	flags := m.LoadFlags()
	*commands = append(*commands, flags...)
}

func loadModules(commands *[]cli.Command) {
	//Load Modules here
	portScan := &modules.Portscan{}
	load(portScan, commands)
	fileControl := &modules.Filecontrol{}
	load(fileControl, commands)
}

func main() {
	app := cli.NewApp()

	var commands []cli.Command

	banner.Init(os.Stdout, true, true, bytes.NewBufferString(b))
	loadModules(&commands)
	app.Commands = commands
	app.Run(os.Args)
}
