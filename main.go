package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
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
		fmt.Println(topicID)
	} else {
		// Lookup the topic ID for use in the publisher
		if topicID, err = client.TopicID(context.Background(), NOAAAlerts); err != nil {
			log.Fatal(err)
		}
	}

	// Run a subscriber
	go func() {
		sub, err := client.Subscribe(context.Background(), topicID)
		if err != nil {
			log.Fatal(err)
		}

		msgs, err := sub.Subscribe()
		if err != nil {
			log.Fatal(err)
		}

		for msg := range msgs {
			fmt.Println(msg)
		}
	}()

	time.Sleep(1 * time.Second)
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
	time.Sleep(1 * time.Second)
}
