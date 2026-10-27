
package model

import (
	"fmt"
	"strings"
	"time"
)

type OrderStatus string

const (
	StatusCreated   OrderStatus = "CREATED"
	StatusConfirmed OrderStatus = "CONFIRMED"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID int64 `json:"id" gorm:"primaryKey;autoIncrement"`

	CustomerID int64 `json:"customerId" gorm:"column:customer_id;not null;index"`
	ProductID  int64 `json:"productId" gorm:"column:product_id;not null;index"`

	Quantity int `json:"quantity" gorm:"not null;check:quantity > 0"`

	UnitPrice float64 `json:"unitPrice" gorm:"column:unit_price;type:numeric(14,2);not null"`
	TotalAmount float64 `json:"totalAmount" gorm:"column:total_amount;type:numeric(14,2);not null"`

	Status OrderStatus `json:"status" gorm:"type:varchar(20);not null;default:CREATED;index"`

	CreatedAt time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}

func (Order) TableName() string {
	return "orders"
}

func (s OrderStatus) IsValid() bool {
	switch s {
	case StatusCreated, StatusConfirmed, StatusCancelled:
		return true
	default:
		return false
	}
}

func ParseOrderStatus(value string) (OrderStatus, error) {
	status := OrderStatus(strings.ToUpper(strings.TrimSpace(value)))

	if !status.IsValid() {
		return "", fmt.Errorf(
			"invalid order status %q: allowed values are CREATED, CONFIRMED, CANCELLED",
			value,
		)
	}

	return status, nil
}

type CreateOrderRequest struct {
	CustomerID int64 `json:"customerId" binding:"required,gt=0"`
	ProductID  int64 `json:"productId" binding:"required,gt=0"`
	Quantity   int   `json:"quantity" binding:"required,gt=0"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
