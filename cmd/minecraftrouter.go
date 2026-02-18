package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/AbandonTech/minecraftrouter/pkg"
	"github.com/AbandonTech/minecraftrouter/pkg/resolver"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
)

func main() {
	var host, configFile, configURL string
	var port uint

	app := &cli.App{
		Name:                 "minecraftrouter",
		Usage:                "route minecraft traffic from a configuration or api",
		Suggest:              true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "host",
				Value:       "0.0.0.0",
				Usage:       "bind listener socket to this host",
				Destination: &host,
				EnvVars:     []string{"ROUTER_HOST"},
			},
			&cli.UintFlag{
				Name:        "port",
				Value:       25565,
				Usage:       "bind listener socket to this port",
				Destination: &port,
				EnvVars:     []string{"ROUTER_PORT"},
			},
			&cli.StringFlag{
				Name:        "file",
				Usage:       "path to a JSON routing config file, e.g. \"routing.json\"",
				Destination: &configFile,
				EnvVars:     []string{"ROUTER_CONFIG_FILE"},
			},
			&cli.StringFlag{
				Name:        "url",
				Usage:       "base URL of the MinecraftAdmin API, e.g. \"https://mcapi.abandontech.cloud/\"",
				Destination: &configURL,
				EnvVars:     []string{"ROUTER_CONFIG_URL"},
			},
			&cli.DurationFlag{
				Name:  "poll-interval",
				Usage: "how often to poll the MinecraftAdmin API for routing config updates",
				Value: 60 * time.Second,
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
			&cli.BoolFlag{
				Name:    "proxy-protocol",
				Usage:   "enable proxy protocol",
				EnvVars: []string{"ROUTER_PROXY_PROTOCOL"},
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

			if configFile != "" && configURL != "" {
				return cli.Exit("--file and --url are mutually exclusive; specify exactly one", 1)
			}
			if configFile == "" && configURL == "" {
				return cli.Exit("one of --file (env: ROUTER_CONFIG_FILE) or --url (env: ROUTER_CONFIG_URL) must be specified", 1)
			}

			var err error
			var resolver_ resolver.Resolver
			if configURL != "" {
				accountID := os.Getenv("MINECRAFT_ADMIN_SERVICE_ACCOUNT_ID")
				secret := os.Getenv("MINECRAFT_ADMIN_SERVICE_ACCOUNT_SECRET")

				if accountID == "" || secret == "" {
					log.Fatal().
						Msg("MINECRAFT_ADMIN_SERVICE_ACCOUNT_ID and MINECRAFT_ADMIN_SERVICE_ACCOUNT_SECRET env vars must be set when using the API resolver.")
				}

				pollInterval := ctx.Duration("poll-interval")
				appCtx := context.Background()

				resolver_, err = resolver.NewApiResolver(appCtx, configURL, accountID, secret, pollInterval)
				if err != nil {
					return err
				}
			} else {
				resolver_, err = resolver.NewJsonResolver(configFile)
				if err != nil {
					return err
				}
			}

			log.Info().
				Str("Host", host).
				Uint("Port", port).
				EmbedObject(resolver_).
				Msg("Configuring router")

			router := pkg.NewRouter(hostAddress, resolver_, ctx.Bool("proxy-protocol"))
			return router.Run()
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().
			Err(err).
			Msg("Error while running application")
	}
}
