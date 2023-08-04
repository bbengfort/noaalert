package noaalert

import (
	"context"
	"time"

	"github.com/rotationalio/go-ensign"
	"github.com/rs/zerolog/log"
)

// EnsureTopicExists checks if the topic exists, and if it doesn't, it creates it.
func EnsureTopicExists(client *ensign.Client, topic string) (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	var exists bool
	if exists, err = client.TopicExists(ctx, topic); err != nil {
		return err
	}
	log.Debug().Bool("exists", exists).Msg("topic exists check")

	if !exists {
		var topicID string
		if topicID, err = client.CreateTopic(ctx, topic); err != nil {
			log.Info().Str("topic", topic).Str("topic_id", topicID).Msg("topic created")
			return err
		}
	}

	return nil
}
