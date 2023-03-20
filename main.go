package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/oklog/ulid"
	ensign "github.com/rotationalio/go-ensign"
	api "github.com/rotationalio/go-ensign/api/v1beta1"
	mimetype "github.com/rotationalio/go-ensign/mimetype/v1beta1"
)

const NOAAAlerts = "noaa-alerts"

func main() {
	// Load environment variables from .env file
	godotenv.Load()

	// Connect to Ensign
	client, err := ensign.New(&ensign.Options{
		ClientID:     os.Getenv("ENSIGN_CLIENT_ID"),
		ClientSecret: os.Getenv("ENSIGN_CLIENT_SECRET"),
		AuthURL:      "https://auth.ensign.world",
		Endpoint:     "ensign.ninja:443",
	})

	if err != nil {
		log.Fatal(err)
	}

	// List topics to check if the topic exists
	exists, err := client.TopicExists(context.Background(), NOAAAlerts)
	if err != nil {
		log.Fatal(err)
	}

	var topicID string
	if !exists {
		if topicID, err = client.CreateTopic(context.Background(), NOAAAlerts); err != nil {
			log.Fatal(err)
		}
	} else {
		// TODO: use a topic lookup method or don't even worry about this
		topics, err := client.ListTopics(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		for _, topic := range topics {
			if topic.Name == NOAAAlerts {
				var topicULID ulid.ULID
				if err = topicULID.UnmarshalBinary(topic.Id); err != nil {
					log.Fatal(err)
				}
				topicID = topicULID.String()
			}
		}

	}

	pub, err := client.Publish(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	data := map[string]interface{}{
		"sender":    "Enson the Sea Otter",
		"timestamp": time.Now().String(),
		"message":   "You're looking smart today!",
	}

	e := &api.Event{
		Mimetype: mimetype.ApplicationJSON,
		Type: &api.Type{
			Name:    "Generic",
			Version: 1,
		},
	}

	e.Data, _ = json.Marshal(data)

	pub.Publish(topicID, e)
}
