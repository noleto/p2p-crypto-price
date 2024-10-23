package cmd

import (
	"context"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/spf13/cobra"
	"log"
)

func listenRun(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	node, err := createNode(ctx)
	defer node.Close()

	topic := joinTopic(err, ctx, node)

	listenCryptoFeed(err, topic, ctx, node)
}

func listenCryptoFeed(err error, topic *pubsub.Topic, ctx context.Context, node host.Host) {
	sub, err := topic.Subscribe()
	if err != nil {
		log.Fatalf("failed to subscribe to topic: %v", err)
	}

	log.Printf("Listening for messages on topic: %s", topicName)

	// Listen for messages
	for {
		msg, err := sub.Next(ctx)
		if err != nil {
			log.Printf("error getting message: %v", err)
			continue
		}

		// Ignore messages from ourselves
		if msg.ReceivedFrom == node.ID() {
			continue
		}

		log.Printf("Received message from %s: %s", msg.ReceivedFrom, string(msg.Data))
	}
}

// listenCmd represents the listen command
var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen to crypto price updates",
	Long:  `Listen to crypto price updates`,
	Run:   listenRun,
}

func init() {
	rootCmd.AddCommand(listenCmd)
}
