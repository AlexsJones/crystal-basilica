package main

import (
	"fmt"
	"go/importer"
	"log"
	"os"
	"reflect"

	r "github.com/AlexsJones/go-type-registry/core"
	"github.com/AlexsJones/gotools/modules"
	runtime "github.com/AlexsJones/gotools/runtime"
	_ "github.com/dimiro1/banner/autoload"
	"github.com/urfave/cli"
)

func loadModule(m runtime.Module, masterCommands *[]cli.Command) {
	moduleCommands := m.LoadFlags()
	*masterCommands = append(*masterCommands, moduleCommands...)
}

func generateRegistry(r *r.Registry) error {
	//Adding modules here
	r.Put(&modules.Gitlab{})
	r.Put(&modules.Jenkins{})
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
		if declName != "Module" && declName != "init" {
			fmt.Println("Loading module: " + declName)
			currentModuleValue, err := registry.Get("*modules." + declName)
			if err != nil {
				fmt.Printf("error: %s\n", err.Error())
				return
			}
			i := currentModuleValue.Unwrap()
			log.Println(reflect.TypeOf(i))
			switch i.(type) {

			case *modules.Gitlab:
				cast := i.(*modules.Gitlab)
				loadModule(cast, &commands)

			case *modules.Jenkins:
				cast := i.(*modules.Jenkins)
				loadModule(cast, &commands)
			}

		}
	}

	app.Commands = commands
	app.Run(os.Args)
}
