package order

import (
	"time"
)

type Order struct {
	ID          string
	CustomerID  string
	Items       []OrderItem
	Status      Status
	TotalAmount int64
	Currency    string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderItem struct {
	ProductID string
	Name      string
	UnitPrice int64
	Quantity  int
}

type Status string

const (
	StatusCreated    Status = "created"
	StatusConfirmed  Status = "confirmed"
	StatusProcessing Status = "processing"
	StatusShipped    Status = "shipped"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)
