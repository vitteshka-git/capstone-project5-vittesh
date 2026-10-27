package tests

import (
	"context"
	"strings"
	"testing"

	"order-management/model"
	"order-management/service"
)

func TestCreateOrderRejectsInvalidCustomerID(t *testing.T) {
	svc := service.NewOrderService(nil, nil)

	req := model.CreateOrderRequest{
		CustomerID: 0,
		ProductID:  101,
		Quantity:   2,
	}

	_, err := svc.CreateOrder(context.Background(), req)

	if err == nil {
		t.Fatal("expected an error for an invalid customer ID")
	}

	if !strings.Contains(err.Error(), "customerId") {
		t.Fatalf("expected customerId validation error, got: %v", err)
	}
}

func TestCreateOrderRejectsInvalidProductID(t *testing.T) {
	svc := service.NewOrderService(nil, nil)

	req := model.CreateOrderRequest{
		CustomerID: 1,
		ProductID:  0,
		Quantity:   2,
	}

	_, err := svc.CreateOrder(context.Background(), req)

	if err == nil {
		t.Fatal("expected an error for an invalid product ID")
	}

	if !strings.Contains(err.Error(), "productId") {
		t.Fatalf("expected productId validation error, got: %v", err)
	}
}

func TestCreateOrderRejectsInvalidQuantity(t *testing.T) {
	svc := service.NewOrderService(nil, nil)

	req := model.CreateOrderRequest{
		CustomerID: 1,
		ProductID:  101,
		Quantity:   0,
	}

	_, err := svc.CreateOrder(context.Background(), req)

	if err == nil {
		t.Fatal("expected an error for an invalid quantity")
	}

	if !strings.Contains(err.Error(), "quantity") {
		t.Fatalf("expected quantity validation error, got: %v", err)
	}
}

func TestGetOrderRejectsInvalidID(t *testing.T) {
	svc := service.NewOrderService(nil, nil)

	_, err := svc.GetOrder(context.Background(), 0)

	if err == nil {
		t.Fatal("expected an error for an invalid order ID")
	}

	if !strings.Contains(err.Error(), "order ID") {
		t.Fatalf("expected order ID validation error, got: %v", err)
	}
}

func TestUpdateOrderStatusRejectsInvalidID(t *testing.T) {
	svc := service.NewOrderService(nil, nil)

	_, err := svc.UpdateOrderStatus(
		context.Background(),
		0,
		"CONFIRMED",
	)

	if err == nil {
		t.Fatal("expected an error for an invalid order ID")
	}
}

func TestUpdateOrderStatusRejectsInvalidStatus(t *testing.T) {
	svc := service.NewOrderService(nil, nil)

	_, err := svc.UpdateOrderStatus(
		context.Background(),
		1,
		"INVALID",
	)

	if err == nil {
		t.Fatal("expected an error for an invalid order status")
	}

	if !strings.Contains(err.Error(), "invalid order status") {
		t.Fatalf("expected invalid status error, got: %v", err)
	}
}

func TestListOrdersRejectsNegativeOffset(t *testing.T) {
	svc := service.NewOrderService(nil, nil)

	_, err := svc.ListOrders(context.Background(), 20, -1)

	if err == nil {
		t.Fatal("expected an error for a negative offset")
	}

	if !strings.Contains(err.Error(), "offset") {
		t.Fatalf("expected offset validation error, got: %v", err)
	}
}
