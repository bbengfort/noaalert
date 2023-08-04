package noaalert

import (
	"os"
	"os/signal"

	sdk "github.com/rotationalio/go-ensign"
	api "github.com/rotationalio/go-ensign/api/v1beta1"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Subscriber struct {
	ensign *sdk.Client
	conf   Config
}

func NewAlerts(conf Config) (sub *Subscriber, err error) {
	if conf.IsZero() {
		if conf, err = NewConfig(); err != nil {
			return nil, err
		}
	}

	if conf.ConsoleLog {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}

	sub = &Subscriber{
		conf: conf,
	}

	if sub.ensign, err = sdk.New(conf.Ensign.Options()...); err != nil {
		return nil, err
	}

	// If we need to ensure the topic exists, perform the check.
	if conf.EnsureTopicExists {
		if err = EnsureTopicExists(sub.ensign, conf.Topic); err != nil {
			return nil, err
		}
	}

	return sub, nil
}

func (s *Subscriber) Run(cb func(*AlertEvent) error) (err error) {
	// Catch OS signals for graceful shutdowns
	var alerts <-chan *AlertEvent
	done := make(chan struct{})
	if alerts, err = s.Listen(done); err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		done <- struct{}{}
	}()

	for alert := range alerts {
		if err = cb(alert); err != nil {
			done <- struct{}{}
			return err
		}
	}
	return nil
}

func (s *Subscriber) Listen(done <-chan struct{}) (_ <-chan *AlertEvent, err error) {
	var sub *sdk.Subscription
	if sub, err = s.ensign.Subscribe(s.conf.Topic); err != nil {
		return nil, err
	}

	alerts := make(chan *AlertEvent, 100)
	go func(alerts chan<- *AlertEvent, sub *sdk.Subscription) {
		log.Info().Str("topic", s.conf.Topic).Msg("listening for alerts")
		defer sub.Close()

		var (
			events  uint64
			skipped uint64
		)

	eventLoop:
		for {
			select {
			case event := <-sub.C:
				log.Debug().Str("id", event.ID()).Str("topic_id", event.TopicID()).Str("type", event.Type.String()).Msg("event recv")

				if event.Type.Name != "Alert" {
					log.Debug().Str("type", event.Type.String()).Msg("unknown type")
					if _, err := event.Nack(api.Nack_UNKNOWN_TYPE); err != nil {
						log.Warn().Err(err).Str("id", event.ID()).Str("reason", "unknown_type").Msg("could not nack event")
					}
					skipped++
					continue eventLoop
				}

				if event.Mimetype != Mimetype {
					log.Debug().Str("mimetype", event.Mimetype.MimeType()).Msg("unknown mimetype")
					if _, err := event.Nack(api.Nack_UNHANDLED_MIMETYPE); err != nil {
						log.Warn().Err(err).Str("id", event.ID()).Str("reason", "unknown_mimetype").Msg("could not nack event")
					}
					skipped++
					continue eventLoop
				}

				alert := &AlertEvent{
					CorrelationID: event.Metadata["correlation_id"],
					RequestID:     event.Metadata["request_id"],
					ServerID:      event.Metadata["server_id"],
					LastModified:  event.Metadata["last_modified"],
					Expires:       event.Metadata["expires"],
					Data:          event.Data,
				}

				if err := alert.parse(); err != nil {
					log.Debug().Err(err).Msg("could not parse alert")
					if _, err := event.Nack(api.Nack_UNPROCESSED); err != nil {
						log.Warn().Err(err).Str("id", event.ID()).Str("reason", "unprocessed").Msg("could not nack event")
					}
					skipped++
					continue eventLoop
				}

				if _, err := event.Ack(); err != nil {
					log.Warn().Err(err).Str("id", event.ID()).Msg("could not ack event")
				}

				alerts <- alert
				events++
			case <-done:
				log.Info().Uint64("events", events).Uint64("skipped", events).Msg("closing subscription channel")
				return
			}
		}

	}(alerts, sub)

	return alerts, nil
}
