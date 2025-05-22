package spannerDb

import "time"

// Order represents the structure for the 'orders' table in Spanner.
type Order struct {
	OrderID   string    `spanner:"order_id"`
	UserID    string    `spanner:"user_id"`
	Amount    float64   `spanner:"amount"`
	Status    string    `spanner:"status"`
	CreatedAt time.Time `spanner:"created_at"`
	UpdatedAt time.Time `spanner:"updated_at"`
}
