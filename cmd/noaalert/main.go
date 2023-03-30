package main

import (
	"fmt"
	"os"

	"github.com/bbengfort/noaalert"
	"github.com/joho/godotenv"
	api "github.com/rotationalio/go-ensign/api/v1beta1"
	"github.com/rs/zerolog/log"
	cli "github.com/urfave/cli/v2"
)

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	app := cli.NewApp()
	app.Name = "noaalert"
	app.Version = "1.0.0"
	app.Usage = "publish NOAA weather alerts to ensign"
	app.Commands = []*cli.Command{
		{
			Name:   "publish",
			Usage:  "run the publisher daemon to fetch alerts from the NOAA API",
			Action: publish,
		},
		{
			Name:   "subscribe",
			Usage:  "subscribe to NOAA alerts on Ensign",
			Action: subscribe,
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

	var events <-chan *api.Event
	if events, err = sub.Listen(); err != nil {
		return cli.Exit(err, 1)
	}

	for event := range events {
		// TODO: do a better job of printing events out
		fmt.Println(event)
	}
	return nil
}
