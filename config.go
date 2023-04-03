package noaalert

import (
	"time"

	"github.com/kelseyhightower/envconfig"
	sdk "github.com/rotationalio/go-ensign"
)

const prefix = "noaalert"

type Config struct {
	ConsoleLog        bool          `split_words:"true" default:"false"`
	EnsureTopicExists bool          `split_words:"true" default:"false"`
	Interval          time.Duration `default:"5m"`
	Ensign            EnsignConfig
	processed         bool
}

type EnsignConfig struct {
	ClientID     string `envconfig:"ENSIGN_CLIENT_ID"`
	ClientSecret string `envconfig:"ENSIGN_CLIENT_SECRET"`
	Endpoint     string `envconfig:"ENSIGN_ENDPOINT"`
	AuthURL      string `envconfig:"ENSIGN_AUTH_URL"`
}

func NewConfig() (conf Config, err error) {
	if err = envconfig.Process(prefix, &conf); err != nil {
		return conf, err
	}

	if err = conf.Validate(); err != nil {
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

func (c EnsignConfig) Options() *sdk.Options {
	opts := sdk.NewOptions()
	if c.ClientID != "" {
		opts.ClientID = c.ClientID
	}

	if c.ClientSecret != "" {
		opts.ClientSecret = c.ClientSecret
	}

	if c.Endpoint != "" {
		opts.Endpoint = c.Endpoint
	}

	if c.AuthURL != "" {
		opts.AuthURL = c.AuthURL
	}

	return opts
}
