package main

import (
	"github.com/AbandonTech/minecraftrouter/pkg"
	"github.com/AbandonTech/minecraftrouter/pkg/resolver"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "minecraftrouter",
		Usage: "route minecraft traffic from a configuration or api",
		Action: func(*cli.Context) error {
			router := pkg.NewRouter("0.0.0.0:25565", resolver.NewApiResolver("http://mcapi.localhost/service/mapping"))
			return router.Run()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
