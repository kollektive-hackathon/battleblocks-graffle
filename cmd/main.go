package main

import (
	"battleblocks-graffle/pubsub"
	"battleblocks-graffle/webhook"

	"github.com/spf13/viper"
)

func main() {
	viper.AutomaticEnv()
	viper.SetDefault("GOOGLE_PROJECT_ID", "flow-battleblocks")
	viper.SetDefault("PUBSUB_TOPIC", "blockchain.flow.events")
	viper.SetDefault("GRAFFLE_COMPANY_ID", "ead2dbd7-47e5-458a-bb65-f2fe4f0dfee2")
	viper.SetDefault("GRAFFLE_SECRET", "3ASDchangemeASD33333")

	project := viper.GetString("PUBSUB_PROJECT")
	topic := viper.GetString("PUBSUB_TOPIC")
	secret := viper.GetString("GRAFFLE_SECRET")

	pubsubClient := pubsub.NewPubsubClient(topic, project)
	webhook := webhook.NewWebhookServer(pubsubClient, secret)
	webhook.Start()
}
