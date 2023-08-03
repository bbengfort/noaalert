package noaalert

import (
	"time"

	"github.com/rotationalio/confire"
	sdk "github.com/rotationalio/go-ensign"
)

const prefix = "noaalert"

type Config struct {
	ConsoleLog        bool          `split_words:"true" default:"false"`
	Topic             string        `default:"noaa-alerts" required:"true"`
	EnsureTopicExists bool          `split_words:"true" default:"false"`
	Interval          time.Duration `default:"5m"`
	Ensign            EnsignConfig
	processed         bool
}

type EnsignConfig struct {
	ClientID     string `env:"ENSIGN_CLIENT_ID" required:"true"`
	ClientSecret string `env:"ENSIGN_CLIENT_SECRET" required:"true"`
	Endpoint     string `env:"ENSIGN_ENDPOINT"`
	AuthURL      string `env:"ENSIGN_AUTH_URL"`
}

func NewConfig() (conf Config, err error) {
	if err = confire.Process(prefix, &conf); err != nil {
		return conf, err
	}

	conf.processed = true
	return conf, nil
}

// A Config is zero-valued if it hasn't been processed by a file or the environment.
func (c Config) IsZero() bool {
	return !c.processed
}

// Mark a manually constructed config as processed as long as its valid.
func (c Config) Mark() (Config, error) {
	if err := c.Validate(); err != nil {
		return c, err
	}
	c.processed = true
	return c, nil
}

// Validates the config is ready for use in the application and that configuration
// semantics such as requiring multiple required configuration parameters are enforced.
func (c Config) Validate() (err error) {
	return nil
}

func (c EnsignConfig) Options() []sdk.Option {
	opts := make([]sdk.Option, 0, 3)
	opts = append(opts, sdk.WithCredentials(c.ClientID, c.ClientSecret))

	if c.Endpoint != "" {
		opts = append(opts, sdk.WithEnsignEndpoint(c.Endpoint, false))
	}

	if c.AuthURL != "" {
		opts = append(opts, sdk.WithAuthenticator(c.AuthURL, false))
	}
	return opts
}
