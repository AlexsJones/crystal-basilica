package main

import (
	"fmt"
	"go/importer"
	"log"
	"os"
	"reflect"
	"strings"

	r "github.com/AlexsJones/go-type-registry/core"
	"github.com/AlexsJones/gotools/modules"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/urfave/cli"
)

func generateRegistry(r *r.Registry) error {
	//Adding modules here
	r.Put(&modules.Portscan{})
	return nil
}

func main() {
	app := cli.NewApp()

	var commands []cli.Command

	//Register types
	registry, err := r.NewRegistry(generateRegistry)
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	//Load modules
	pkg, err := importer.Default().Import("github.com/AlexsJones/gotools/modules")
	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return
	}
	for _, declName := range pkg.Scope().Names() {
		currentType := pkg.Scope().Lookup(declName).Type().String()
		if !strings.Contains(currentType, "github.com/AlexsJones/gotools/modules") {
			continue
		}

		fmt.Println("Loading module: " + declName)
		currentModuleValue, err := registry.Get("*modules." + declName)
		if err != nil {
			fmt.Printf("error: %s\n", err.Error())
			return
		}
		i := currentModuleValue.Unwrap()
		log.Println(reflect.TypeOf(i))

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
