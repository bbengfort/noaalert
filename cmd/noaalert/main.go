package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/bbengfort/noaalert"
	"github.com/joho/godotenv"
	confire "github.com/rotationalio/confire/usage"
	"github.com/rotationalio/go-ensign"
	api "github.com/rotationalio/go-ensign/api/v1beta1"
	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
)

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	app := cli.NewApp()
	app.Name = "noaalert"
	app.Version = noaalert.Version()
	app.Usage = "publish NOAA weather alerts to Ensign"
	app.Commands = []*cli.Command{
		{
			Name:     "publish",
			Usage:    "run the publisher daemon to fetch alerts from the NOAA API",
			Category: "server",
			Action:   publish,
		},
		{
			Name:     "info",
			Usage:    "fetch project info stats and usage",
			Category: "utility",
			Action:   projectInfo,
		},
		{
			Name:     "subscribe",
			Category: "utility",
			Usage:    "subscribe to NOAA alerts on Ensign",
			Action:   subscribe,
		},
		{
			Name:     "alerts",
			Category: "utility",
			Usage:    "get active NOAA alerts",
			Action:   alerts,
		},
		{
			Name:     "config",
			Usage:    "print noaalerts configuration guide",
			Category: "utility",
			Action:   usage,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "list",
					Aliases: []string{"l"},
					Usage:   "print in list mode instead of table mode",
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal().Err(err).Msg("could not execute cli app")
	}
}

func publish(c *cli.Context) (err error) {
	var conf noaalert.Config
	if conf, err = noaalert.NewConfig(); err != nil {
		return cli.Exit(err, 1)
	}

	var pub *noaalert.Publisher
	if pub, err = noaalert.New(conf); err != nil {
		return cli.Exit(err, 1)
	}

	if err = pub.Run(); err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func subscribe(c *cli.Context) (err error) {
	var conf noaalert.Config
	if conf, err = noaalert.NewConfig(); err != nil {
		return cli.Exit(err, 1)
	}

	var sub *noaalert.Subscriber
	if sub, err = noaalert.NewAlerts(conf); err != nil {
		return cli.Exit(err, 1)
	}

	err = sub.Run(func(alert *noaalert.AlertEvent) (err error) {
		var headline string
		if headline, err = alert.Headline(); err != nil {
			log.Warn().Err(err).Msg("could not get headline from alert")
			return nil
		}

		log.Info().Msg(headline)
		return nil
	})

	if err != nil {
		return cli.Exit(err, 1)
	}
	return nil
}

func alerts(c *cli.Context) (err error) {
	var api *noaalert.Weather
	if api, err = noaalert.NewWeatherAPI(); err != nil {
		return cli.Exit(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var events []*noaalert.AlertEvent
	if events, err = api.Alerts(ctx); err != nil {
		return cli.Exit(err, 1)
	}

	for _, event := range events {
		var headline string
		if headline, err = event.Headline(); err != nil {
			continue
		}
		fmt.Println(headline)
	}
	return nil
}

func usage(c *cli.Context) (err error) {
	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, 4, ' ', 0)
	format := confire.DefaultTableFormat
	if c.Bool("list") {
		format = confire.DefaultListFormat
	}

	var conf noaalert.Config
	if err := confire.Usagef("noaalert", &conf, tabs, format); err != nil {
		return cli.Exit(err, 1)
	}
	tabs.Flush()
	return nil
}

func projectInfo(c *cli.Context) (err error) {
	var conf noaalert.Config
	if conf, err = noaalert.NewConfig(); err != nil {
		return cli.Exit(err, 1)
	}

	var client *ensign.Client
	if client, err = ensign.New(conf.Ensign.Options()...); err != nil {
		return cli.Exit(err, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var info *api.ProjectInfo
	if info, err = client.Info(ctx); err != nil {
		return cli.Exit(err, 1)
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err = encoder.Encode(info); err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}
