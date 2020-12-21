package main

import (
	"context"
	"fmt"
	"github.com/glours/docker-lint/internal"
	"os"

	"github.com/docker/cli/cli-plugins/manager"
	"github.com/docker/cli/cli-plugins/plugin"
	"github.com/docker/cli/cli/command"
	"github.com/spf13/cobra"
)

func main() {
	ctx, closeFunc := context.WithCancel(context.Background())
	defer closeFunc()
	plugin.Run(func(dockerCli command.Cli) *cobra.Command {

		return newLintCmd(ctx, dockerCli)
	},
		manager.Metadata{
			SchemaVersion: "0.1.0",
			Vendor:        "Docker Inc.",
			Version:       internal.Version,
		})
}

type options struct {
	showVersion bool
}

func newLintCmd(ctx context.Context, cli command.Cli) *cobra.Command {
	var flags options
	cmd := &cobra.Command{
		Short:       "Docker Lint",
		Long:        `A tool to lint your Dockerfiles and Compose files`,
		Use:         "lint [OPTIONS] DOCKERFILE",
		RunE: func(cmd *cobra.Command, args []string) error {
			if flags.showVersion {
				return runVersion()
			}
			if len(args) > 0 {
				delegate("hadolint", os.Args[2:]...)
			} else {
				if err := cmd.Usage(); err !=nil {
					return err
				}
				return fmt.Errorf(`"docker lint" requires at least 1 argument`)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&flags.showVersion, "version", false, "Display version of the lint plugin")
	return cmd
}

func runVersion() error {
	version, err := internal.FullVersion()
	if err != nil {
		return err
	}
	fmt.Println(version)
	return nil
}
