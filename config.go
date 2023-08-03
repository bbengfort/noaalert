package noaalert

import (
	"fmt"
	"strings"
	"time"

	"github.com/rotationalio/confire"
	sdk "github.com/rotationalio/go-ensign"
	"github.com/rs/zerolog"
)

const prefix = "noaalert"

type Config struct {
	Topic             string        `default:"noaa-alerts" required:"true"`
	EnsureTopicExists bool          `split_words:"true" default:"false"`
	Interval          time.Duration `default:"5m" required:"true"`
	ConsoleLog        bool          `split_words:"true" default:"false"`
	LogLevel          LevelDecoder  `default:"info" split_words:"true"`
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

// Parse and return the zerolog log level for configuring global logging.
func (c Config) GetLogLevel() zerolog.Level {
	return zerolog.Level(c.LogLevel)
}

// LogLevelDecoder deserializes the log level from a config string.
type LevelDecoder zerolog.Level

// Names of log levels for use in encoding/decoding from strings.
const (
	llPanic = "panic"
	llFatal = "fatal"
	llError = "error"
	llWarn  = "warn"
	llInfo  = "info"
	llDebug = "debug"
	llTrace = "trace"
)

// Decode implements confire Decoder interface.
func (ll *LevelDecoder) Decode(value string) error {
	value = strings.TrimSpace(strings.ToLower(value))
	switch value {
	case llPanic:
		*ll = LevelDecoder(zerolog.PanicLevel)
	case llFatal:
		*ll = LevelDecoder(zerolog.FatalLevel)
	case llError:
		*ll = LevelDecoder(zerolog.ErrorLevel)
	case llWarn:
		*ll = LevelDecoder(zerolog.WarnLevel)
	case llInfo:
		*ll = LevelDecoder(zerolog.InfoLevel)
	case llDebug:
		*ll = LevelDecoder(zerolog.DebugLevel)
	case llTrace:
		*ll = LevelDecoder(zerolog.TraceLevel)
	default:
		return fmt.Errorf("unknown log level %q", value)
	}
	return nil
}
