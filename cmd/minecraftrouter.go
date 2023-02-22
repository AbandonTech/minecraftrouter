package main

import (
	"github.com/AbandonTech/minecraftrouter/pkg"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "minecraftrouter",
		Usage: "route minecraft traffic from a configuration or api",
		Action: func(*cli.Context) error {
			_, err := pkg.NewResolver("routing.json")
			if err != nil {
				return err
			}
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
