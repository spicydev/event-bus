package pubsubMsg

import "time"

// OrderEvent matches the structure expected by genericMessageReceiver for "order" type messages.
type OrderEvent struct {
	EventTimestamp time.Time `json:"event_timestamp"`
	Data           OrderData `json:"data"`
}

type OrderData struct {
	OrderID string  `json:"order_id"`
	UserID  string  `json:"user_id"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
}
