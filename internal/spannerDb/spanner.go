package spannerDb

import (
	"context"
	"fmt"

	"cloud.google.com/go/spanner"
	"github.com/spicydev/event-bus/v2/config"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	spannerClient *spanner.Client
}

func NewClient(ctx context.Context, props *config.Spanner) (*Client, error) {
	var client *spanner.Client
	var err error
	databaseName := fmt.Sprintf("projects/%s/instances/%s/databases/%s",
		props.ProjectID, props.InstanceID, props.DatabaseID)

	if emulatorAddr := props.EmulatorHost; props.EmulatorEnabled && emulatorAddr != "" {
		emulatorOpts := []option.ClientOption{
			option.WithEndpoint(emulatorAddr),
			option.WithGRPCDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
			option.WithoutAuthentication(),
			internaloption.SkipDialSettingsValidation(),
		}
		client, err = spanner.NewClient(ctx, databaseName, emulatorOpts...)
	} else {
		client, err = spanner.NewClient(ctx, databaseName, option.WithCredentialsFile(props.CredentialsFile))
	}
	if err != nil {
		return nil, err
	}
	return &Client{spannerClient: client}, nil
}

func (c *Client) Close() {
	c.spannerClient.Close()
}

type Repository[T any] struct {
	spannerClient *spanner.Client
	table         string
}

func NewRepository[T any](client *Client, table string) *Repository[T] {
	return &Repository[T]{spannerClient: client.spannerClient, table: table}
}

func (r *Repository[T]) Save(ctx context.Context, entity T) error {
	mutation, err := spanner.InsertOrUpdateStruct(r.table, entity)
	if err != nil {
		return err
	}
	_, err = r.spannerClient.Apply(ctx, []*spanner.Mutation{mutation})
	return err
}

func (r *Repository[T]) SaveAll(ctx context.Context, entities []T) error {
	var mutations []*spanner.Mutation
	for _, entity := range entities {
		m, _ := spanner.InsertOrUpdateStruct(r.table, entity)
		mutations = append(mutations, m)
	}
	_, err := r.spannerClient.Apply(ctx, mutations)
	return err
}
