package noaalert

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"time"

	sdk "github.com/rotationalio/go-ensign"
	api "github.com/rotationalio/go-ensign/api/v1beta1"
	mimetype "github.com/rotationalio/go-ensign/mimetype/v1beta1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	Topic = "noaa-alerts"
)

func init() {
	// Initializes zerolog with our default logging requirements
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.DurationFieldInteger = false
	zerolog.DurationFieldUnit = time.Millisecond
}

func TopicID(client *sdk.Client, create bool) (topicID string, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var exists bool
	if exists, err = client.TopicExists(ctx, Topic); err != nil {
		return "", err
	}

	if !exists {
		if create {
			return client.CreateTopic(ctx, Topic)
		}
		return "", fmt.Errorf("topic %q does not exist", Topic)
	}
	return client.TopicID(ctx, Topic)
}

type Publisher struct {
	api     *Weather
	ensign  *sdk.Client
	conf    Config
	started time.Time
	echan   chan error
}

func New(conf Config) (pub *Publisher, err error) {
	if conf.IsZero() {
		if conf, err = NewConfig(); err != nil {
			return nil, err
		}
	}

	if conf.ConsoleLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	pub = &Publisher{
		conf:  conf,
		echan: make(chan error, 1),
	}

	// Connect to Weather.gov
	if pub.api, err = NewWeatherAPI(); err != nil {
		return nil, err
	}

	// Connect to Ensign
	if pub.ensign, err = sdk.New(conf.Ensign.Options()); err != nil {
		return nil, err
	}

	return pub, nil
}

func (p *Publisher) Run() error {
	// Catch OS signals for graceful shutdowns
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		p.echan <- p.Shutdown()
	}()

	p.started = time.Now()
	ticker := time.NewTicker(p.conf.Interval)

	topicID, err := TopicID(p.ensign, p.conf.EnsureTopicExists)
	if err != nil {
		return err
	}

	pub, err := p.ensign.Publish(context.Background())
	if err != nil {
		return err
	}

	// Begin API query loop
	for {
		select {
		case err := <-p.echan:
			return err
		case <-ticker.C:
			count := 0
			for _, alert := range p.Alerts() {
				pub.Publish(topicID, &api.Event{Data: alert, Mimetype: mimetype.ApplicationJSON})
				count++
			}
			log.Info().Int("count", count).Msg("weather alerts published")
		}
	}
}

func (p *Publisher) Shutdown() (err error) {
	return nil
}

func (p *Publisher) Alerts() [][]byte {
	// TODO: set default timeout in configuration
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	alerts, err := p.api.Alerts(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("could not fetch noaa alerts")
		return nil
	}

	data := make([][]byte, 0, len(alerts))
	for _, alert := range alerts {
		// TODO: don't republish alerts
		alertjson, err := json.Marshal(alert)
		if err != nil {
			log.Warn().Err(err).Msg("could not parse alert json")
			continue
		}
		data = append(data, alertjson)
	}

	return data
}
