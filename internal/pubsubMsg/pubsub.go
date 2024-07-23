package pubsubMsg

import (
	"context"
	"fmt"

	"cloud.google.com/go/pubsub"
	"github.com/spicydev/event-bus/v2/config"
	"github.com/spicydev/event-bus/v2/internal/spannerDb"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
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

func (c *Client) StartMsgReceivers(ctx context.Context, props config.PubSub, client *spannerDb.Client) {
	accountSubName := props.SubscriptionID
	sub := c.pubsubClient.Subscription(accountSubName)
	AccountReceiver(sub, ctx, accountSubName, client)
}
