package pubsubMsg

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// AccountEvent matches the structure expected by genericMessageReceiver for "account" type messages.
type AccountEvent struct {
	EventTimestamp time.Time `json:"event_timestamp"`
	Data           AccountData `json:"data"`
}

type AccountData struct {
	AccountID string  `json:"account_id"`
	Name      string  `json:"name"`
	Balance   float64 `json:"balance"`
}
