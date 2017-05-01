package modules

import (
	"errors"
	"fmt"
	"os"

	"github.com/pierrre/archivefile/zip"
	"github.com/urfave/cli"
)

//Filecontrol ...
type Filecontrol struct {
}

//LoadFlags for cli
func (p *Filecontrol) LoadFlags() []cli.Command {

	var commands []cli.Command = make([]cli.Command, 0)
	n := cli.Command{
		Name:    "file",
		Aliases: []string{"f"},
		Usage:   "boring file operations condensed",
		Subcommands: []cli.Command{
			{
				Name:    "zip",
				Aliases: []string{"z"},
				Usage:   "Please either a path to a file or folder to zip & output <PATH> <OUTPUT_PATH>",
				Action: func(c *cli.Context) error {

					path := c.Args().Get(0)
					outpath := c.Args().Get(1)
					if path == "" || outpath == "" {
						errMessage := "Please either a path to a file or folder to zip & output <PATH> <OUTPUT_PATH>"
						fmt.Println(errMessage)
						return errors.New(errMessage)
					}
					fi, err := os.Stat(path)
					if err != nil {
						fmt.Println(err)
						return errors.New(err.Error())
					}
					progress := func(archivePath string) {
						fmt.Println(archivePath)
					}
					outFilePath, err := os.Create(outpath)
					if err != nil {
						fmt.Println(err)
						return errors.New(err.Error())
					}
					defer outFilePath.Close()

					switch mode := fi.Mode(); {
					case mode.IsDir():
						err = zip.Archive(path, outFilePath, progress)
						if err != nil {
							fmt.Println(err)
							return errors.New(err.Error())
						}
					case mode.IsRegular():
						err = zip.ArchiveFile(path, outpath, progress)
						if err != nil {
							fmt.Println(err)
							return errors.New(err.Error())
						}
					}
					return nil
				},
			},
		},
	}

	commands = append(commands, n)
	return commands
}
