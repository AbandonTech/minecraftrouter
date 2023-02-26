package main

import (
	"fmt"
	"github.com/AbandonTech/minecraftrouter/pkg"
	"github.com/AbandonTech/minecraftrouter/pkg/resolver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
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
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "verbose log output",
				Aliases: []string{"v"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "pretty",
				Usage:   "pretty log output",
				Aliases: []string{"p"},
				Value:   false,
			},
		},
		Before: func(context *cli.Context) error {
			zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

			zerolog.SetGlobalLevel(zerolog.InfoLevel)
			if context.Bool("verbose") {
				zerolog.SetGlobalLevel(zerolog.DebugLevel)
			}

			if context.Bool("pretty") {
				log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
			}

			log.Debug().
				Bool("Verbose", context.Bool("verbose")).
				Bool("Pretty", context.Bool("pretty")).
				Msg("Configured logging")
			return nil
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

			log.Info().
				Str("Host", host).
				Uint("Port", port).
				EmbedObject(resolver_).
				Msg("Configuring router")

			router := pkg.NewRouter(hostAddress, resolver_)
			return router.Run()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().
			Err(err).
			Msg("error while running application")
	}
}
