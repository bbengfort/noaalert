package noaalert

import (
	"os"

	sdk "github.com/rotationalio/go-ensign"
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

func (s *Subscriber) Listen() (*sdk.Subscription, error) {
	return s.ensign.Subscribe(s.conf.Topic)
}
