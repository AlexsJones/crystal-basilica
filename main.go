package main

import (
	"bytes"
	"fmt"
	"go/importer"
	"os"
	"strings"

	r "github.com/AlexsJones/go-type-registry/core"
	"github.com/AlexsJones/schism/modules"
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

func generateRegistry(r *r.Registry) error {
	//Adding modules here
	r.Put(&modules.Portscan{})
	return nil
}

func main() {
	app := cli.NewApp()
	var commands []cli.Command

	banner.Init(os.Stdout, true, true, bytes.NewBufferString(b))
	//Register types
	registry, err := r.NewRegistry(generateRegistry)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	//Load modules
	pkg, err := importer.Default().Import("github.com/AlexsJones/schism/modules")
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	for _, declName := range pkg.Scope().Names() {
		currentType := pkg.Scope().Lookup(declName).Type().String()
		if !strings.Contains(currentType, "github.com/AlexsJones/schism/modules") {
			continue
		}

		currentModuleValue, err := registry.Get("*modules." + declName)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			return
		}
		i := currentModuleValue.Unwrap()

		switch i.(type) {

		case *modules.Portscan:
			ps := i.(*modules.Portscan)
			moduleCommands := ps.LoadFlags()
			commands = append(commands, moduleCommands...)
		}
	}

	app.Commands = commands
	app.Run(os.Args)
}
