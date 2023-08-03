package noaalert

import (
	"context"
	"time"

	"github.com/rotationalio/go-ensign"
)

// EnsureTopicExists checks if the topic exists, and if it doesn't, it creates it.
func EnsureTopicExists(client *ensign.Client, topic string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var exists bool
	if exists, err = client.TopicExists(ctx, topic); err != nil {
		return err
	}

	if !exists {
		if _, err = client.CreateTopic(ctx, topic); err != nil {
			return err
		}
	}

	return nil
}
