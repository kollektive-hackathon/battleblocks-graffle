package main

import (
	"battleblocks-graffle/pubsub"
	"battleblocks-graffle/webhook"

	"github.com/spf13/viper"
)

func main() {
	viper.SetDefault("PUBSUB_PROJECT", "battleblocks-test")
	viper.SetDefault("PUBSUB_TOPIC", "blockchain.flow.events")

	project := viper.GetString("PUBSUB_PROJECT")
	topic := viper.GetString("PUBSUB_TOPIC")

	pubsubClient := pubsub.NewPubsubClient(topic, project)
	webhook := webhook.NewWebhookServer(pubsubClient)
	webhook.Start()
}
