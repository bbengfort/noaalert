package noaalert

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rotationalio/go-ensign"
	api "github.com/rotationalio/go-ensign/api/v1beta1"
)

func Query(client *ensign.Client, offset, limit int) (_ *AlertIterator, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var cursor *ensign.QueryCursor
	if cursor, err = client.EnSQL(ctx, &api.Query{Query: fmt.Sprintf("SELECT * FROM noaa-alerts OFFSET %d LIMIT %d", offset, limit)}); err != nil {
		return nil, err
	}

	return &AlertIterator{cursor: cursor}, nil
}

type AlertIterator struct {
	err     error
	done    bool
	current *AlertEvent
	cursor  *ensign.QueryCursor
}

func (a *AlertIterator) Next() bool {
	if a.done {
		return false
	}

	event, err := a.cursor.FetchOne()
	if err != nil {
		if !errors.Is(err, ensign.ErrNoRows) {
			a.err = err
		}
		a.done = true
		return false
	}

	a.current = &AlertEvent{
		CorrelationID: event.Metadata["correlation_id"],
		RequestID:     event.Metadata["request_id"],
		ServerID:      event.Metadata["server_id"],
		LastModified:  event.Metadata["last_modified"],
		Expires:       event.Metadata["expires"],
		Data:          event.Data,
	}

	return true
}

func (a *AlertIterator) Alert() *AlertEvent {
	return a.current
}

func (a *AlertIterator) Error() error {
	return a.err
}

func (a *AlertIterator) Release() {
	if err := a.cursor.Close(); err != nil {
		a.err = errors.Join(a.err, err)
	}
}
