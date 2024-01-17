package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/cluttrdev/cli"
)

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
    root := &cli.Command{
        Name: "hello",
        ShortHelp: "Say hello to the world.",
        ShortUsage: "hello <name>",
        Exec: func(ctx context.Context, args []string) error {
            if len(args) < 1 {
                return errors.New("not enough arguments")
            } else if len(args) > 1 {
                return errors.New("too many arguments")
            }

            _, err := fmt.Printf("Hello, %s.\n", args[0])
            return err
        },
    }

    return root
}

