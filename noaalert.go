package noaalert

import (
	"context"
	"os"
	"os/signal"
	"time"

	sdk "github.com/rotationalio/go-ensign"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func init() {
	// Initializes zerolog with our default logging requirements
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.DurationFieldInteger = false
	zerolog.DurationFieldUnit = time.Millisecond
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
	if pub.ensign, err = sdk.New(conf.Ensign.Options()...); err != nil {
		return nil, err
	}

	// If we need to ensure the topic exists, perform the check.
	if conf.EnsureTopicExists {
		if err = EnsureTopicExists(pub.ensign, conf.Topic); err != nil {
			return nil, err
		}
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
	log.Info().Dur("interval", p.conf.Interval).Str("topic", p.conf.Topic).Msg("starting alerts publisher")

	// Begin API query loop
queryLoop:
	for {
		select {
		case err := <-p.echan:
			return err
		case <-ticker.C:
			log.Debug().Msg("starting collection of noaa alerts")

			count := 0
			for alert := range p.Alerts() {
				if err := p.ensign.Publish(p.conf.Topic, alert); err != nil {
					log.Error().Err(err).Int("count", count).Msg("could not publish weather alert")
					continue queryLoop
				}

				// TODO: check acks/nacks
				count++
			}
			log.Info().Str("topic", p.conf.Topic).Int("count", count).Msg("weather alerts published")
		}
	}
}

func (p *Publisher) Shutdown() (err error) {
	log.Info().Msg("shutting alert publisher down")
	if err = p.ensign.Close(); err != nil {
		return err
	}
	log.Debug().Msg("gracefully shut down alert publisher")
	return nil
}

func (p *Publisher) Alerts() <-chan *sdk.Event {
	events := make(chan *sdk.Event)
	go func(events chan<- *sdk.Event) {
		defer close(events)

		// TODO: set default timeout in configuration
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		alerts, err := p.api.Alerts(ctx)
		if err != nil {
			log.Warn().Err(err).Msg("could not fetch noaa alerts")
			return
		}

		log.Debug().Int("nalerts", len(alerts)).Msg("received alerts from NOAA")
		for _, alert := range alerts {
			// TODO: don't republish alerts
			events <- alert.Event()
		}
	}(events)
	return events
}
