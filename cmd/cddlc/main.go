package main

import (
	"log"
	"os"

	"github.com/HannesKimara/cddlc/cmd/cddlc/commands"
	"github.com/urfave/cli/v2"
)

var (
	version = ""
	commit  = ""
	date    = ""
)

func main() {
	app := &cli.App{
		Name:  "cddl",
		Usage: "CDDL validator and code generator",
		Commands: []*cli.Command{
			{
				Name:   "init",
				Usage:  "Initialize a new cddlc project",
				Action: commands.InitCmd,
			},
			{
				Name:    "generate",
				Usage:   "Generate code from definition file",
				Aliases: []string{"gen"},
				Action:  commands.GenerateCmd,
			},
			{
				Name:   "lex",
				Usage:  "Export tokens from cddl source code",
				Action: LexerCmd,
				Hidden: true,
			},
			{
				Name:   "parse",
				Usage:  "Export AST representation of the source code",
				Action: ParseCmd,
				Hidden: true,
			},
			{
				Name:   "doctor",
				Usage:  "Show information about the current build",
				Action: DoctorCmd,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}