package pubsubMsg

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/spicydev/event-bus/v2/config"
	"github.com/spicydev/event-bus/v2/internal/spannerDb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"encoding/json"
	"log"
	"sync"

	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	pubsubClient *pubsub.Client
}

func NewClient(ctx context.Context, cfg *config.PubSub) (*Client, error) {
	var client *pubsub.Client
	var err error
	if emulatorAddr := cfg.EmulatorHost; cfg.EmulatorEnabled && cfg.EmulatorHost != "" {
		conn, e := grpc.Dial(emulatorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if e != nil {
			return nil, fmt.Errorf("grpc.Dial: %w", e)
		}
		emulatorOpts := []option.ClientOption{
			option.WithGRPCConn(conn),
			option.WithTelemetryDisabled(),
		}
		client, err = pubsub.NewClient(ctx, cfg.ProjectID, emulatorOpts...)
	} else {
		client, err = pubsub.NewClient(ctx, cfg.ProjectID, option.WithCredentialsFile(cfg.CredentialsFile))
	}
	if err != nil {
		return nil, err
	}
	return &Client{pubsubClient: client}, nil
}

func (c *Client) Close() {
	c.pubsubClient.Close()
}

func (c *Client) StartMsgReceivers(ctx context.Context, props config.PubSub, spannerClient *spannerDb.Client) {
	var wg sync.WaitGroup
	for _, subConfig := range props.Subscriptions {
		wg.Add(1)
		go func(sc config.SubscriptionConfig) {
			defer wg.Done()
			sub := c.pubsubClient.Subscription(sc.ID)
			log.Printf("Starting receiver for subscription %s, entity type %s, target table %s", sc.ID, sc.EntityType, sc.TargetTable)
			genericMessageReceiver(sub, ctx, sc, spannerClient)
		}(subConfig)
	}
	wg.Wait()
	log.Println("All message receivers stopped.")
}

func genericMessageReceiver(sub *pubsub.Subscription, ctx context.Context, subConfig config.SubscriptionConfig, spannerClient *spannerDb.Client) {
	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		log.Printf("INFO: [Subscription: %s, MessageID: %s] Received message. Data: %s", subConfig.ID, msg.ID, string(msg.Data))

		var err error
		processed := false

		switch subConfig.EntityType {
		case "account":
			var event AccountEvent
			if err = json.Unmarshal(msg.Data, &event); err != nil {
				log.Printf("ERROR: [Subscription: %s, MessageID: %s] Error unmarshalling account event: %v", subConfig.ID, msg.ID, err)
				msg.Nack() // Nack the message if unmarshalling fails
				return
			}
			account := spannerDb.Account{
				AccountID: event.Data.AccountID,
				Name:      event.Data.Name,
				Balance:   event.Data.Balance,
				CreatedAt: event.EventTimestamp,
				UpdatedAt: event.EventTimestamp,
			}
			repo := spannerDb.NewRepository[spannerDb.Account](spannerClient, subConfig.TargetTable)
			if err = repo.Save(ctx, account); err != nil {
				log.Printf("ERROR: [Subscription: %s, MessageID: %s] Error saving account to Spanner: %v", subConfig.ID, msg.ID, err)
				// Decide on Nack or Ack based on error type, for now, Nack for all errors
				msg.Nack()
				return
			}
			processed = true
		case "order":
			var event OrderEvent
			if err = json.Unmarshal(msg.Data, &event); err != nil {
				log.Printf("ERROR: [Subscription: %s, MessageID: %s] Error unmarshalling order event: %v", subConfig.ID, msg.ID, err)
				msg.Nack()
				return
			}
			order := spannerDb.Order{
				OrderID:   event.Data.OrderID,
				UserID:    event.Data.UserID,
				Amount:    event.Data.Amount,
				Status:    event.Data.Status,
				CreatedAt: event.EventTimestamp,
				UpdatedAt: event.EventTimestamp,
			}
			repo := spannerDb.NewRepository[spannerDb.Order](spannerClient, subConfig.TargetTable)
			if err = repo.Save(ctx, order); err != nil {
				log.Printf("ERROR: [Subscription: %s, MessageID: %s] Error saving order to Spanner: %v", subConfig.ID, msg.ID, err)
				msg.Nack()
				return
			}
			processed = true
		default:
			log.Printf("WARN: [Subscription: %s, MessageID: %s] Unknown entity type '%s'. Message will be acked without processing.", subConfig.ID, msg.ID, subConfig.EntityType)
			// Ack the message as we don't know how to process it, to avoid it being redelivered indefinitely.
			// Alternatively, Nack if there should be a dead-letter queue or other handling for unknown types.
		}

		if processed {
			log.Printf("INFO: [Subscription: %s, MessageID: %s] Successfully processed message for entity type '%s'. Acking.", subConfig.ID, msg.ID, subConfig.EntityType)
			msg.Ack()
		} else {
			// If not processed due to unknown type, Ack to remove from subscription.
			// If processing failed for a known type, it should have been Nacked already.
			log.Printf("INFO: [Subscription: %s, MessageID: %s] Message (entity type '%s') was not processed (e.g. unknown type), acking.", subConfig.ID, msg.ID, subConfig.EntityType)
			msg.Ack()
		}
	})
	if err != nil {
		// This error is generally only non-nil if context is cancelled or there's a critical subscription issue.
		log.Printf("ERROR: [Subscription: %s] Error from Receive loop: %v. Receiver will stop.", subConfig.ID, err)
	}
}
