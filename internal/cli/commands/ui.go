package commands

import (
	"github.com/terzigolu/josepshbrain-go/internal/tui"
	"github.com/urfave/cli/v2"
)

// NewUICommand creates the 'ui' command which launches the TUI
func NewUICommand() *cli.Command {
	return &cli.Command{
		Name:  "ui",
		Usage: "Launch interactive terminal UI",
		Action: func(c *cli.Context) error {
			return tui.Run()
		},
	}
}
