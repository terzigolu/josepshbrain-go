package main

import (
	"log"
	"os"

	"github.com/terzigolu/josepshbrain-go/internal/cli/commands"
	"github.com/urfave/cli/v2"
)

// Version will be set during build with ldflags
var Version = "dev"

func main() {
	app := &cli.App{
		Name:    "jbraincli",
		Usage:   "A CLI for interacting with the JosephsBrain API",
		Version: Version,
		Commands: []*cli.Command{
			commands.NewSetupCommand(),
			commands.NewTaskCommand(),
			commands.NewProjectCommand(),
			commands.NewMemoryCommand(),
			commands.NewRememberCommand(), // Direct remember command
			commands.NewKanbanCmd(),
			commands.NewAnnotateCmd(),
			commands.NewTaskAnnotationsCmd(),
			commands.NewContextCommand(),
			commands.NewConfigCommand(),
			commands.NewGeminiKeyCommand(),
			commands.NewAnalyzeCommand(),
			commands.NewUICommand(),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
