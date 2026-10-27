package product

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"order-management/config"
)

var ErrCircuitOpen = errors.New("product service circuit breaker is open")

type PriceResponse struct {
	ProductID int64   `json:"productId"`
	Price     float64 `json:"price"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	maxRetries int
	breaker    *CircuitBreaker
}

func NewClient(cfg *config.Config) (*Client, error) {
	baseURL := strings.TrimRight(cfg.ProductServiceURL, "/")

	parsedURL, err := url.Parse(baseURL)
	if err != nil || parsedURL.Host == "" ||
		(parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		return nil, errors.New("invalid Product Service URL")
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: cfg.ProductServiceTimeout,
		},
		maxRetries: cfg.ProductServiceMaxRetries,
		breaker:    NewCircuitBreaker(5, 30*time.Second),
	}, nil
}

// GetPrice retrieves the price for a product.
// The endpoint contract used here is GET /api/v1/products/{id}.
func (c *Client) GetPrice(
	ctx context.Context,
	productID int64,
) (float64, error) {
	if productID <= 0 {
		return 0, errors.New("product ID must be greater than zero")
	}

	if err := c.breaker.Allow(); err != nil {
		return 0, err
	}

	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if err := ctx.Err(); err != nil {
			c.breaker.RecordFailure()
			return 0, err
		}

		price, err := c.fetchPrice(ctx, productID)
		if err == nil {
			c.breaker.RecordSuccess()
			return price, nil
		}

		lastErr = err

		if !isRetryable(err) || attempt == c.maxRetries {
			break
		}

		// Exponential backoff: 100ms, 200ms, 400ms, ...
		delay := 100 * time.Millisecond * time.Duration(
			1<<uint(attempt),
		)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			c.breaker.RecordFailure()
			return 0, ctx.Err()
		case <-timer.C:
		}
	}

	c.breaker.RecordFailure()
	return 0, fmt.Errorf("Product Service request failed: %w", lastErr)
}

func (c *Client) fetchPrice(
	ctx context.Context,
	productID int64,
) (float64, error) {
	endpoint := c.baseURL + "/api/v1/products/" +
		strconv.FormatInt(productID, 10)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("create Product Service request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("call Product Service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests ||
		resp.StatusCode >= http.StatusInternalServerError {
		return 0, &retryableError{
			err: fmt.Errorf("Product Service returned HTTP %d", resp.StatusCode),
		}
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf(
			"Product Service returned HTTP %d",
			resp.StatusCode,
		)
	}

	var result PriceResponse
	if err := json.NewDecoder(
		io.LimitReader(resp.Body, 1<<20),
	).Decode(&result); err != nil {
		return 0, fmt.Errorf("decode Product Service response: %w", err)
	}

	if result.ProductID != productID {
		return 0, errors.New("Product Service returned a mismatched product ID")
	}

	if math.IsNaN(result.Price) ||
		math.IsInf(result.Price, 0) ||
		result.Price < 0 {
		return 0, errors.New("Product Service returned an invalid price")
	}

	return result.Price, nil
}

type retryableError struct {
	err error
}

func (e *retryableError) Error() string {
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

func isRetryable(err error) bool {
	var retryErr *retryableError
	if errors.As(err, &retryErr) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return errors.Is(err, context.DeadlineExceeded)
}
