package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"
)

type PubsubClient struct {
	topic pubsub.Topic;
}

func NewPubsubClient(topic, project string) *PubsubClient {
	client, err := pubsub.NewClient(context.Background(), project)
	if err != nil {
		panic("cannot initialize pubsub client")
	}

	return &PubsubClient{
		topic: *client.Topic(topic),
	}
}

func (pc *PubsubClient) Publish(ctx context.Context, message *pubsub.Message) {
	log.Info().Interface("message", message.Data).Msg("Sending graffle event message to topic")
	pc.topic.Publish(ctx, message)
}
