package product

import (
	"sync"
	"time"
)

type breakerState int

const (
	stateClosed breakerState = iota
	stateOpen
	stateHalfOpen
)

type CircuitBreaker struct {
	mu               sync.Mutex
	state            breakerState
	failures         int
	failureThreshold int
	recoveryTimeout  time.Duration
	openedAt         time.Time
	probeInFlight    bool
}

func NewCircuitBreaker(
	failureThreshold int,
	recoveryTimeout time.Duration,
) *CircuitBreaker {
	if failureThreshold < 1 {
		failureThreshold = 5
	}

	if recoveryTimeout <= 0 {
		recoveryTimeout = 30 * time.Second
	}

	return &CircuitBreaker{
		state:            stateClosed,
		failureThreshold: failureThreshold,
		recoveryTimeout:  recoveryTimeout,
	}
}

// Allow determines whether a request may call the Product Service.
func (cb *CircuitBreaker) Allow() error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case stateClosed:
		return nil

	case stateOpen:
		if time.Since(cb.openedAt) < cb.recoveryTimeout {
			return ErrCircuitOpen
		}

		// Recovery time has passed. Permit one test request.
		cb.state = stateHalfOpen
		cb.probeInFlight = true
		return nil

	case stateHalfOpen:
		// Only one test request may run at a time.
		if cb.probeInFlight {
			return ErrCircuitOpen
		}

		cb.probeInFlight = true
		return nil

	default:
		return ErrCircuitOpen
	}
}

// RecordSuccess closes the circuit after a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0
	cb.state = stateClosed
	cb.probeInFlight = false
}

// RecordFailure opens the circuit when the failure threshold is reached.
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.probeInFlight = false

	if cb.state == stateHalfOpen {
		cb.state = stateOpen
		cb.openedAt = time.Now()
		return
	}

	cb.failures++

	if cb.failures >= cb.failureThreshold {
		cb.state = stateOpen
		cb.openedAt = time.Now()
	}
}
