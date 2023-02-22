package main

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "minecraftrouter",
		Usage: "route minecraft traffic from a configuration or api",
		Action: func(*cli.Context) error {
			fmt.Println("Hello world")
			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
