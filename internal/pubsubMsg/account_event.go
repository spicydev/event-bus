package pubsubMsg

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/spicydev/event-bus/v2/internal/spannerDb"
)

type AccountEvent struct {
	UserId         int64     `json:"id"`
	UserGuid       string    `json:"guid"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	Phone          string    `json:"phone"`
	Email          string    `json:"email"`
	CreatedTs      time.Time `json:"createdTs"`
	LastModifiedTs time.Time `json:"lastModifiedTs"`
}

func AccountReceiver(subscription *pubsub.Subscription, ctx context.Context, subName string, client *spannerDb.Client) {
	go func() {
		fmt.Printf("Listening for messages on Pub/Sub subscription: %v \n", subName)
		err := subscription.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			handleMessage(ctx, msg, client)
		})
		if err != nil {
			log.Fatalf("Error receiving messages: %v", err)
		}
	}()
}

func handleMessage(ctx context.Context, msg *pubsub.Message, client *spannerDb.Client) {
	var payload AccountEvent
	if err := json.Unmarshal(msg.Data, &payload); err != nil {
		log.Printf("Error unmarshalling Pub/Sub message: %v", err)
		msg.Nack()
		return
	}
	fmt.Printf("Received message with ID: %v", msg.ID)
	var xrefs []spannerDb.AccountXref
	xrefs = append(xrefs, spannerDb.AccountXref{
		XrefVal:        payload.Email,
		XrefType:       "EMAIL",
		UserGuid:       payload.UserGuid,
		UserId:         payload.UserId,
		CreatedTs:      payload.CreatedTs,
		LastModifiedTs: payload.LastModifiedTs,
	}, spannerDb.AccountXref{
		XrefVal:        payload.Phone,
		XrefType:       "PHONE",
		UserGuid:       payload.UserGuid,
		UserId:         payload.UserId,
		CreatedTs:      payload.CreatedTs,
		LastModifiedTs: payload.LastModifiedTs,
	})
	account := spannerDb.Account{
		UserGuid:       payload.UserGuid,
		UserId:         payload.UserId,
		FirstName:      payload.FirstName,
		LastName:       payload.LastName,
		Phone:          payload.Phone,
		Email:          payload.Email,
		CreatedTs:      payload.CreatedTs,
		LastModifiedTs: payload.LastModifiedTs,
	}
	xrefRepo := spannerDb.NewRepository[spannerDb.AccountXref](client, "USER_XREF")
	err := xrefRepo.SaveAll(ctx, xrefs)
	if err != nil {
		fmt.Printf("Error saving xref events to db: %v \n", err)
	}
	accRepo := spannerDb.NewRepository[spannerDb.Account](client, "USER")
	error := accRepo.Save(ctx, account)
	if error != nil {
		fmt.Printf("Error saving user event to db: %v \n", error)
	}
	if err == nil && error == nil {
		fmt.Println("Sucessfully Persisted User account to Spanner")
		msg.Ack()
	}
}
