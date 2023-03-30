package noaalert

import (
	"context"
	"os"

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

	if sub.ensign, err = sdk.New(conf.Ensign.Options()); err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *Subscriber) Listen() (<-chan *api.Event, error) {
	topicID, err := TopicID(s.ensign, s.conf.EnsureTopicExists)
	if err != nil {
		return nil, err
	}

	stream, err := s.ensign.Subscribe(context.Background(), topicID)
	if err != nil {
		return nil, err
	}

	return stream.Subscribe()
}
