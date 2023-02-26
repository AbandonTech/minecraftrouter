package main

import (
	"fmt"
	"github.com/AbandonTech/minecraftrouter/pkg"
	"github.com/AbandonTech/minecraftrouter/pkg/resolver"
	"github.com/urfave/cli/v2"
	"log"
	url2 "net/url"
	"os"
)

func main() {
	var host, lookup string
	var port uint

	app := &cli.App{
		Name:                 "minecraftrouter",
		Usage:                "route minecraft traffic from a configuration or api",
		Suggest:              true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "host",
				Value:       "127.0.0.1",
				Usage:       "bind listener socket to this host",
				Destination: &host,
				EnvVars:     []string{"MINECRAFT_ROUTER_HOST"},
			},
			&cli.UintFlag{
				Name:        "port",
				Value:       25565,
				Usage:       "bind listener socket to this port",
				Destination: &port,
				EnvVars:     []string{"MINECRAFT_ROUTER_PORT"},
			},
			&cli.StringFlag{
				Name: "lookup",
				Usage: "lookup file or api to use for routing, " +
					"for example \"routing.json\" or \"http://localhost:8002/service/mapping\"",
				Destination: &lookup,
				EnvVars:     []string{"MINECRAFT_ROUTER_LOOKUP"},
				Required:    true,
			},
		},
		Action: func(ctx *cli.Context) error {
			hostAddress := fmt.Sprintf("%s:%d", host, port)

			url, err := url2.Parse(lookup)
			if err != nil {
				return err
			}

			var resolver_ resolver.Resolver
			if url.Scheme == "http" || url.Scheme == "https" {
				resolver_ = resolver.NewApiResolver(lookup)
			} else {
				resolver_, err = resolver.NewJsonResolver(lookup)
				if err != nil {
					return err
				}
			}

			router := pkg.NewRouter(hostAddress, resolver_)
			return router.Run()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
