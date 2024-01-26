package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/cluttrdev/cli"
)

const version string = "v0.0.0+unknown"

func main() {
	if err := execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func execute() error {
	cmd := configure()

	args := os.Args[1:]
	opts := []cli.ParseOption{}

	if err := cmd.Parse(args, opts...); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		} else {
			return fmt.Errorf("error parsing arguments: %w", err)
		}
	}

	return cmd.Run(context.Background())
}

func configure() *cli.Command {
	out := os.Stdout

	root := &cli.Command{
		Name:       "version",
		ShortUsage: "version <command>",
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
	}

	// default version info taken from BuildInfo
	defaultCmd := cli.DefaultVersionCommand(out)
	defaultCmd.Name = "default"
	defaultCmd.ShortHelp = "Show default version information"

	// custom version info based on BuildInfo but explicit version number
	customInfo := cli.NewBuildInfo(version)
	customCmd := cli.NewVersionCommand(customInfo, out)
	customCmd.Name = "custom"
	customCmd.ShortHelp = "Show custom version information"

	root.Subcommands = []*cli.Command{
		defaultCmd,
		customCmd,
	}

	return root
}
