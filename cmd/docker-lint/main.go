package main

import (
	"os"

	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

func main() {
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "lint",
			Short: "A basic Hello World plugin",
			RunE: func(cmd *cobra.Command, args []string) error {
				delegate("hadolint", os.Args[2:]...)
				return nil
			},
		}

		return cmd
	},
		manager.Metadata{
			SchemaVersion: "0.1.0",
			Vendor:        "Docker Inc.",
			Version:       "testing",
		})
}
