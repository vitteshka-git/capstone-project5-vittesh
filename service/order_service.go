package service

import (
	"context"
	"errors"
	"fmt"
	"math"

	"order-management/model"
	"order-management/repository"
)

type ProductPriceProvider interface {
	GetPrice(ctx context.Context, productID int64) (float64, error)
}

type OrderService struct {
	repo           *repository.OrderRepository
	productService ProductPriceProvider
}

func NewOrderService(
	repo *repository.OrderRepository,
	productService ProductPriceProvider,
) *OrderService {
	return &OrderService{
		repo:           repo,
		productService: productService,
	}
}

func (s *OrderService) CreateOrder( ctx context.Context, req model.CreateOrderRequest,) (*model.Order, error) {
	if req.CustomerID <= 0 {
		return nil, errors.New("customerId must be greater than zero")
	}
	if req.ProductID <= 0 {
		return nil, errors.New("productId must be greater than zero")
	}
	if req.Quantity <= 0 {
		return nil, errors.New("quantity must be greater than zero")
	}
	if s.productService == nil {
		return nil, errors.New("product service is not configured")
	}

	unitPrice, err := s.productService.GetPrice(ctx, req.ProductID)
	if err != nil {
		return nil, fmt.Errorf("retrieve product price: %w", err)
	}
	if math.IsNaN(unitPrice) || math.IsInf(unitPrice, 0) || unitPrice < 0 {
		return nil, errors.New("product service returned an invalid price")
	}

	unitPrice = math.Round(unitPrice*100) / 100
	totalAmount := unitPrice * float64(req.Quantity)

	if math.IsInf(totalAmount, 0) ||
		math.IsNaN(totalAmount) ||
		totalAmount > 999999999999.99 {
		return nil, errors.New("calculated order total is out of range")
	}

	order := &model.Order{
		CustomerID:  req.CustomerID,
		ProductID:   req.ProductID,
		Quantity:    req.Quantity,
		UnitPrice:   unitPrice,
		TotalAmount: math.Round(totalAmount*100) / 100,
		Status:      model.StatusCreated,
	}

	if s.repo == nil {
		return nil, errors.New("order repository is not configured")
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("save order: %w", err)
	}

	return order, nil
}

func (s *OrderService) GetOrder(
	ctx context.Context,
	id int64,
) (*model.Order, error) {
	if id <= 0 {
		return nil, errors.New("order ID must be greater than zero")
	}
	if s.repo == nil {
		return nil, errors.New("order repository is not configured")
	}

	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("retrieve order: %w", err)
	}

	return order, nil
}

func (s *OrderService) ListOrders(
	ctx context.Context,
	limit, offset int,
) ([]model.Order, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		return nil, errors.New("offset cannot be negative")
	}
	if s.repo == nil {
		return nil, errors.New("order repository is not configured")
	}

	orders, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}

	return orders, nil
}

func (s *OrderService) UpdateOrderStatus(
	ctx context.Context,
	id int64,
	statusText string,
) (*model.Order, error) {
	if id <= 0 {
		return nil, errors.New("order ID must be greater than zero")
	}

	status, err := model.ParseOrderStatus(statusText)
	if err != nil {
		return nil, err
	}
	if s.repo == nil {
		return nil, errors.New("order repository is not configured")
	}

	order, err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, fmt.Errorf("update order status: %w", err)
	}

	return order, nil
}
