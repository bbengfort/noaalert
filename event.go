package noaalert

import (
	"encoding/json"

	"github.com/rotationalio/go-ensign"
	api "github.com/rotationalio/go-ensign/api/v1beta1"
	mimetype "github.com/rotationalio/go-ensign/mimetype/v1beta1"
)

type AlertEvent struct {
	CorrelationID string
	RequestID     string
	ServerID      string
	LastModified  string
	Expires       string
	Data          []byte
	parsed        map[string]interface{}
}

var Mimetype = mimetype.ApplicationJSON

var AlertType = &api.Type{
	Name:         "Alert",
	MajorVersion: 1,
	MinorVersion: 0,
	PatchVersion: 0,
}

func (a *AlertEvent) Event() *ensign.Event {
	meta := make(ensign.Metadata)
	meta["correlation_id"] = a.CorrelationID
	meta["request_id"] = a.RequestID
	meta["server_id"] = a.ServerID
	meta["last_modified"] = a.LastModified
	meta["expires"] = a.Expires

	return &ensign.Event{
		Metadata: meta,
		Data:     a.Data,
		Type:     AlertType,
		Mimetype: Mimetype,
	}
}

func (a *AlertEvent) Headline() (_ string, err error) {
	if err = a.parse(); err != nil {
		return "", err
	}

	props, ok := a.parsed["properties"].(map[string]interface{})
	if !ok {
		return "", ErrNoProperties
	}

	headline, ok := props["headline"].(string)
	if !ok {
		return "", ErrNoHeadline
	}
	return headline, nil
}

func (a *AlertEvent) parse() error {
	if a.parsed == nil {
		return json.Unmarshal(a.Data, &a.parsed)
	}
	return nil
}
