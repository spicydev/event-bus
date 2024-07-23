package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spicydev/event-bus/v2/config"
	"github.com/spicydev/event-bus/v2/internal/pubsubMsg"
	"github.com/spicydev/event-bus/v2/internal/spannerDb"
)

func main() {
	props := loadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup Spanner Client Connection
	spanner_client := loadSpannerClient(ctx, props)
	defer spanner_client.Close()

	// Setup PubSub Client Connection
	pubsubClient := loadPubSubClient(ctx, props)
	defer pubsubClient.Close()
	pubsubClient.StartMsgReceivers(ctx, props.PubSub, spanner_client)

	// Setup a channel to receive OS signals for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	// Wait for a termination signal
	<-signals
	log.Println("INFO: Shutdown signal received, shutting down gracefully...")

	// Cancel the context to stop receiving messages
	cancel()

	// Allow some time for in-flight requests to complete
	time.Sleep(time.Duration(props.Server.ShutdownDelay) * time.Second)
	log.Println("INFO: Server shut down successfully")
}

func loadSpannerClient(ctx context.Context, props *config.Props) *spannerDb.Client {
	client, err := spannerDb.NewClient(ctx, &props.Spanner)
	if err != nil {
		log.Fatalf("ERROR: Failed to create Spanner client: %v", err)
	}
	return client
}

func loadPubSubClient(ctx context.Context, props *config.Props) *pubsubMsg.Client {
	client, err := pubsubMsg.NewClient(ctx, &props.PubSub)
	if err != nil {
		log.Fatalf("ERROR: Failed to create PubSub Client: %v", err)
	}
	return client
}

func loadConfig() *config.Props {
	props, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("ERROR: Error loading config: %v", err)
	}
	return props
}
