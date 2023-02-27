package pubsub

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/rs/zerolog/log"
)

type PubsubClient struct {
	topic pubsub.Topic
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
	res := pc.topic.Publish(ctx, message)
	go func(res *pubsub.PublishResult) {
		_, err := res.Get(ctx)
		if err != nil {
			log.Warn().Msg(fmt.Sprintf("Failed to publish message: %s", err.Error()))
			return
		}
		log.Info().Msg("Sending successfull")

	}(res)
}
