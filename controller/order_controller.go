package controller

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"order-management/model"
	"order-management/repository"
	"order-management/service"

	"github.com/gin-gonic/gin"
)

type OrderController struct {
	orderService *service.OrderService
}

func NewOrderController(orderService *service.OrderService) *OrderController {
	return &OrderController{
		orderService: orderService,
	}
}

// POST /api/v1/orders
func (c *OrderController) CreateOrder(ctx *gin.Context) {
	var req model.CreateOrderRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request body: customerId, productId, and quantity must be valid positive numbers",
		})
		return
	}

	order, err := c.orderService.CreateOrder(ctx.Request.Context(), req)
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{
			"error": "Unable to create order",
		})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"message": "Order created successfully",
		"data":    order,
	})
}

// GET /api/v1/orders/:id
func (c *OrderController) GetOrder(ctx *gin.Context) {
	id, err := parseOrderID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	order, err := c.orderService.GetOrder(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to retrieve order",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": order})
}

// GET /api/v1/orders
func (c *OrderController) ListOrders(ctx *gin.Context) {
	limit := 20
	offset := 0

	if value := ctx.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "limit must be a positive integer",
			})
			return
		}
		limit = parsed
	}

	if value := ctx.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "offset must be a non-negative integer",
			})
			return
		}
		offset = parsed
	}

	if limit > 100 {
		limit = 100
	}

	orders, err := c.orderService.ListOrders(
		ctx.Request.Context(),
		limit,
		offset,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to list orders",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":   orders,
		"limit":  limit,
		"offset": offset,
	})
}

// PUT /api/v1/orders/:id/status
func (c *OrderController) UpdateOrderStatus(ctx *gin.Context) {
	id, err := parseOrderID(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req model.UpdateOrderStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Request body must contain a valid status",
		})
		return
	}

	status := strings.ToUpper(strings.TrimSpace(req.Status))
	if _, err := model.ParseOrderStatus(status); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Status must be CREATED, CONFIRMED, or CANCELLED",
		})
		return
	}

	order, err := c.orderService.UpdateOrderStatus(
		ctx.Request.Context(),
		id,
		status,
	)
	if err != nil {
		if errors.Is(err, repository.ErrOrderNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
			return
		}

		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Unable to update order status",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
		"data":    order,
	})
}

func parseOrderID(value string) (int64, error) {
	id, err := strconv.ParseInt(value, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("order ID must be a positive integer")
	}
	return id, nil
}
