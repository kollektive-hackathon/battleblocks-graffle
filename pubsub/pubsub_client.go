package pubsub

import (
	"context"

	"cloud.google.com/go/pubsub"
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
	pc.topic.Publish(ctx, message)
}
