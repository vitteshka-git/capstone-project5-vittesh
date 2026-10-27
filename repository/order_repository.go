
package repository

import (
	"context"
	"errors"
	"fmt"

	"order-management/model"

	"gorm.io/gorm"
)

var ErrOrderNotFound = errors.New("order not found")

type OrderRepository struct {
	DB *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{DB: db}
}

// Create saves a new order.
func (r *OrderRepository) Create(
	ctx context.Context,
	order *model.Order,
) error {
	if err := r.DB.WithContext(ctx).Create(order).Error; err != nil {
		return fmt.Errorf("create order: %w", err)
	}

	return nil
}

// GetByID retrieves one order.
func (r *OrderRepository) GetByID(
	ctx context.Context,
	id int64,
) (*model.Order, error) {
	var order model.Order

	err := r.DB.WithContext(ctx).First(&order, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get order: %w", err)
	}

	return &order, nil
}

// List retrieves orders with pagination.
func (r *OrderRepository) List(
	ctx context.Context,
	limit, offset int,
) ([]model.Order, error) {
	var orders []model.Order

	err := r.DB.WithContext(ctx).
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&orders).Error

	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}

	return orders, nil
}

// UpdateStatus changes an order's status.
func (r *OrderRepository) UpdateStatus(
	ctx context.Context,
	id int64,
	status model.OrderStatus,
) (*model.Order, error) {
	result := r.DB.WithContext(ctx).
		Model(&model.Order{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return nil, fmt.Errorf("update order status: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return nil, ErrOrderNotFound
	}

	return r.GetByID(ctx, id)
}
