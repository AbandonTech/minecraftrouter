package main

import (
	"fmt"
	mcProtocol "github.com/AbandonTech/minecraftrouter/pkg/minecraft_protocol"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v2"
	"net"
	"os"
)

func main() {
	var host string
	var port uint

	app := &cli.App{
		Name:                 "minecraftstats",
		Usage:                "gather metrics from a minecraft server",
		Suggest:              true,
		EnableBashCompletion: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "host",
				Value:       "127.0.0.1",
				Usage:       "minecraft host url to gather metrics from",
				Destination: &host,
				EnvVars:     []string{"MINECRAFT_ROUTER_HOST"},
			},
			&cli.UintFlag{
				Name:        "port",
				Value:       25565,
				Usage:       "minecraft port to gather metrics from",
				Destination: &port,
				EnvVars:     []string{"MINECRAFT_ROUTER_PORT"},
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
			address := fmt.Sprintf("%s:%d", host, port)

			conn, err := net.Dial("tcp", address)
			if err != nil {
				return err
			}

			client := mcProtocol.NewReadWriter(conn)
			client.Write([]byte{1, 2, 3, 4})

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().
			Err(err).
			Msg("Error while running application")
	}
}
