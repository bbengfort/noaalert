package noaalert_test

import (
	"os"
	"testing"
	"time"

	"github.com/bbengfort/noaalert"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var testEnv = map[string]string{
	"NOAALERT_TOPIC":               "testing-alerts",
	"NOAALERT_ENSURE_TOPIC_EXISTS": "true",
	"NOAALERT_INTERVAL":            "1m",
	"NOAALERT_LOG_LEVEL":           "warn",
	"NOAALERT_CONSOLE_LOG":         "true",
	"ENSIGN_CLIENT_ID":             "abcdefg1234",
	"ENSIGN_CLIENT_SECRET":         "abcdefghijklmnopqrstuvwxyz1234567",
	"ENSIGN_ENDPOINT":              "localhost:8000",
	"ENSIGN_AUTH_URL":              "http://localhost:8001",
}

func TestConfig(t *testing.T) {
	// Set required environment variables and cleanup after the test is complete.
	t.Cleanup(cleanupEnv())
	setEnv()

	conf, err := noaalert.NewConfig()
	require.NoError(t, err, "could not process configuration from the environment")
	require.False(t, conf.IsZero(), "processed config should not be zero valued")

	// Ensure configuration is correctly set from the environment
	require.Equal(t, testEnv["NOAALERT_TOPIC"], conf.Topic)
	require.True(t, conf.EnsureTopicExists)
	require.Equal(t, 1*time.Minute, conf.Interval)
	require.Equal(t, zerolog.WarnLevel, conf.GetLogLevel())
	require.Equal(t, testEnv["ENSIGN_CLIENT_ID"], conf.Ensign.ClientID)
	require.Equal(t, testEnv["ENSIGN_CLIENT_SECRET"], conf.Ensign.ClientSecret)
	require.Equal(t, testEnv["ENSIGN_ENDPOINT"], conf.Ensign.Endpoint)
	require.Equal(t, testEnv["ENSIGN_AUTH_URL"], conf.Ensign.AuthURL)
	require.True(t, conf.ConsoleLog)
}

func TestOptions(t *testing.T) {
	// Set required environment variables and cleanup after the test is complete.
	t.Cleanup(cleanupEnv())
	setEnv()

	conf, err := noaalert.NewConfig()
	require.NoError(t, err, "could not process configuration from the environment")

	opts := conf.Ensign.Options()
	require.Len(t, opts, 3)
}

func TestLevelDecoder(t *testing.T) {
	testTable := []struct {
		value    string
		expected zerolog.Level
	}{
		{
			"panic", zerolog.PanicLevel,
		},
		{
			"FATAL", zerolog.FatalLevel,
		},
		{
			"Error", zerolog.ErrorLevel,
		},
		{
			"   warn   ", zerolog.WarnLevel,
		},
		{
			"iNFo", zerolog.InfoLevel,
		},
		{
			"debug", zerolog.DebugLevel,
		},
		{
			"trace", zerolog.TraceLevel,
		},
	}

	// Test valid cases
	for _, testCase := range testTable {
		var level noaalert.LevelDecoder
		err := level.Decode(testCase.value)
		require.NoError(t, err)
		require.Equal(t, testCase.expected, zerolog.Level(level))
	}

	// Test error case
	var level noaalert.LevelDecoder
	err := level.Decode("notalevel")
	require.EqualError(t, err, `unknown log level "notalevel"`)
}

// Returns the current environment for the specified keys, or if no keys are specified
// then it returns the current environment for all keys in the testEnv variable.
func curEnv(keys ...string) map[string]string {
	env := make(map[string]string)
	if len(keys) > 0 {
		for _, key := range keys {
			if val, ok := os.LookupEnv(key); ok {
				env[key] = val
			}
		}
	} else {
		for key := range testEnv {
			env[key] = os.Getenv(key)
		}
	}

	return env
}

// Sets the environment variables from the testEnv variable. If no keys are specified,
// then this function sets all environment variables from the testEnv.
func setEnv(keys ...string) {
	if len(keys) > 0 {
		for _, key := range keys {
			if val, ok := testEnv[key]; ok {
				os.Setenv(key, val)
			}
		}
	} else {
		for key, val := range testEnv {
			os.Setenv(key, val)
		}
	}
}

// Cleanup helper function that can be run when the tests are complete to reset the
// environment back to its previous state before the test was run.
func cleanupEnv(keys ...string) func() {
	prevEnv := curEnv(keys...)
	return func() {
		for key, val := range prevEnv {
			if val != "" {
				os.Setenv(key, val)
			} else {
				os.Unsetenv(key)
			}
		}
	}
}
