package noaalert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	UserAgent      = "(bbengfort.github.io, benjamin@bengfort.com)"
	BaseWeatherURL = "https://api.weather.gov"
)

type Weather struct {
	client  *http.Client
	baseURL *url.URL
}

func NewWeatherAPI() (api *Weather, err error) {
	api = &Weather{
		client: &http.Client{
			Transport:     nil,
			CheckRedirect: nil,
			Timeout:       30 * time.Second,
		},
	}

	if api.client.Jar, err = cookiejar.New(nil); err != nil {
		return nil, fmt.Errorf("could not create cookiejar: %w", err)
	}

	if api.baseURL, err = url.Parse(BaseWeatherURL); err != nil {
		return nil, err
	}

	return api, nil
}

func (s *Weather) Alerts(ctx context.Context) (_ []*AlertEvent, err error) {
	var req *http.Request
	if req, err = s.NewRequest(ctx, http.MethodGet, "/alerts/active", nil, nil); err != nil {
		return nil, err
	}

	var rep *http.Response
	alerts := make(map[string]interface{})
	if rep, err = s.Do(req, &alerts, true); err != nil {
		logctx := log.With().Err(err).Logger()
		if rep != nil {
			// Get the NOAA request headers to log the error
			correlationID := rep.Header.Get("X-Correlation-Id")
			requestID := rep.Header.Get("X-Request-Id")
			serverID := rep.Header.Get("X-Server-Id")

			logctx = logctx.With().
				Str("correlation_id", correlationID).
				Str("request_id", requestID).
				Str("server_id", serverID).
				Logger()
		}
		logctx.Error().Msg("could not fetch active alerts")
		return nil, err
	}

	if features, ok := alerts["features"]; ok {
		if featureList, ok := features.([]interface{}); ok {
			// Get the NOAA request headers to create events
			correlationID := rep.Header.Get("X-Correlation-Id")
			requestID := rep.Header.Get("X-Request-Id")
			serverID := rep.Header.Get("X-Server-Id")
			lastModified := rep.Header.Get("Last-Modified")
			expires := rep.Header.Get("Expires")

			events := make([]*AlertEvent, 0, len(featureList))
			for _, feature := range featureList {
				event := &AlertEvent{
					CorrelationID: correlationID,
					RequestID:     requestID,
					ServerID:      serverID,
					LastModified:  lastModified,
					Expires:       expires,
				}

				if event.Data, err = json.Marshal(feature); err != nil {
					return nil, err
				}

				events = append(events, event)
			}

			return events, nil
		}
	}
	return nil, fmt.Errorf("no alerts returned")
}

const (
	accept      = "application/geo+json"
	acceptLang  = "en-US,en"
	contentType = "application/json; charset=utf-8"
)

func (s *Weather) NewRequest(ctx context.Context, method, path string, data interface{}, params *url.Values) (req *http.Request, err error) {
	// Resolve the URL reference from the path
	url := s.baseURL.ResolveReference(&url.URL{Path: path})
	if params != nil && len(*params) > 0 {
		url.RawQuery = params.Encode()
	}

	var body io.ReadWriter
	switch {
	case data == nil:
		body = nil
	default:
		body = &bytes.Buffer{}
		if err = json.NewEncoder(body).Encode(data); err != nil {
			return nil, fmt.Errorf("could not serialize request data as json: %s", err)
		}
	}

	// Create the http request
	if req, err = http.NewRequestWithContext(ctx, method, url.String(), body); err != nil {
		return nil, fmt.Errorf("could not create request: %s", err)
	}

	// Set the headers on the request
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Accept", accept)
	req.Header.Add("Accept-Language", acceptLang)

	if body != nil {
		req.Header.Add("Content-Type", contentType)
	}

	// Add CSRF protection if its available
	if s.client.Jar != nil {
		cookies := s.client.Jar.Cookies(url)
		for _, cookie := range cookies {
			if cookie.Name == "csrf_token" {
				req.Header.Add("X-CSRF-TOKEN", cookie.Value)
			}
		}
	}

	return req, nil
}

// Do executes an http request against the server, performs error checking, and
// deserializes the response data into the specified struct.
func (s *Weather) Do(req *http.Request, data interface{}, checkStatus bool) (rep *http.Response, err error) {
	if rep, err = s.client.Do(req); err != nil {
		return rep, fmt.Errorf("could not execute request: %s", err)
	}
	defer rep.Body.Close()

	// Detect http status errors if they've occurred
	if checkStatus {
		if rep.StatusCode < 200 || rep.StatusCode >= 300 {
			return rep, fmt.Errorf("[%d] %s", rep.StatusCode, rep.Status)
		}
	}

	// Deserialize the JSON data from the body
	if data != nil && rep.StatusCode >= 200 && rep.StatusCode < 300 && rep.StatusCode != http.StatusNoContent {
		// Check the content type to ensure data deserialization is possible
		if ct := rep.Header.Get("Content-Type"); ct != accept {
			return rep, fmt.Errorf("unexpected content type: %q", ct)
		}

		if err = json.NewDecoder(rep.Body).Decode(data); err != nil {
			return nil, fmt.Errorf("could not deserialize response data: %s", err)
		}
	}

	return rep, nil
}

func (s *Weather) SetBaseURL(u *url.URL) {
	s.baseURL = u
}
