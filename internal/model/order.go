package model

import (
	"time"
)

type Order struct {
	ID         int         `json:"id"`
	UserID     int         `json:"user_id"`
	Status     string      `json:"status"`
	TotalPrice float64     `json:"total_price"`
	CreatedAt  time.Time   `json:"created_at"`
	Items      []OrderItem `json:"items,omitempty"`
}

type CreateOrderRequest struct {
	UserID int               `json:"user_id"`
	Items  []CreateOrderItem `json:"items"`
}

type CreateOrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}
