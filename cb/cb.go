package cb

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

const (
	Closed   = "closed"
	Open     = "open"
	HalfOpen = "half-open"
)

// circuitBreaker manages the state and behavior of the circuit breaker
type circuitBreaker struct {
	mu                   sync.Mutex
	state                string
	failureCount         int
	lastFailureTime      time.Time
	halfOpenSuccessCount int

	failureThreshold    int
	recoveryTime        time.Duration
	halfOpenMaxRequests int
}

// NewCircuitBreaker initializes a new CircuitBreaker
func NewCircuitBreaker(failureThreshold int, recoveryTime time.Duration, halfOpenMaxRequests int) *circuitBreaker {
	return &circuitBreaker{
		state:               Closed,
		failureThreshold:    failureThreshold,
		recoveryTime:        recoveryTime,
		halfOpenMaxRequests: halfOpenMaxRequests,
	}
}

// Call attempts to execute the provided function, managing state transitions
func (cb *circuitBreaker) Call(fn func() (any, error)) (any, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	slog.Info("Making a request", "state", cb.state)

	switch cb.state {
	case Closed:
		return cb.handleClosedState(fn)
	case Open:
		return cb.handleOpenState()
	case HalfOpen:
		return cb.handleHalfOpenState(fn)
	default:
		return nil, errors.New("unknown circuit state")
	}
}

// handleClosedState executes the function and monitors failures
func (cb *circuitBreaker) handleClosedState(fn func() (any, error)) (any, error) {
	result, err := cb.runWithTimeout(fn)
	if err != nil {
		slog.Warn("Request failed in closed state", "failureCount", cb.failureCount+1)
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		if cb.failureCount >= cb.failureThreshold {
			cb.state = Open
			slog.Error("Failure threshold reached, transitioning to open")
		}
		return nil, err
	}

	slog.Info("Request succeeded in closed state")
	cb.resetCircuit()
	return result, nil
}

// handleOpenState blocks requests if recovery time hasn't passed
func (cb *circuitBreaker) handleOpenState() (any, error) {
	if time.Since(cb.lastFailureTime) > cb.recoveryTime {
		cb.state = HalfOpen
		cb.halfOpenSuccessCount = 0
		cb.failureCount = 0
		slog.Info("Recovery period over, transitioning to half-open")
		return nil, nil
	}

	slog.Warn("Circuit is still open, blocking request")
	return nil, errors.New("circuit open, request blocked")
}

// handleHalfOpenState executes the function and checks for recovery
func (cb *circuitBreaker) handleHalfOpenState(fn func() (any, error)) (any, error) {
	result, err := cb.runWithTimeout(fn)
	if err != nil {
		slog.Error("Request failed in half-open state, transitioning to open")
		cb.state = Open
		cb.lastFailureTime = time.Now()
		return nil, err
	}

	cb.halfOpenSuccessCount++
	slog.Info("Request succeeded in half-open state", "successCount", cb.halfOpenSuccessCount)

	if cb.halfOpenSuccessCount >= cb.halfOpenMaxRequests {
		slog.Info("Max success in half-open, transitioning to closed")
		cb.resetCircuit()
	}

	return result, nil
}

// runWithTimeout executes the provided function with a timeout
func (cb *circuitBreaker) runWithTimeout(fn func() (any, error)) (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resultChan := make(chan struct {
		result any
		err    error
	}, 1)

	go func() {
		result, err := fn()
		resultChan <- struct {
			result any
			err    error
		}{result, err}
	}()

	select {
	case <-ctx.Done():
		return nil, errors.New("request timed out")
	case res := <-resultChan:
		return res.result, res.err
	}
}

// resetCircuit resets the circuit breaker to closed state
func (cb *circuitBreaker) resetCircuit() {
	cb.failureCount = 0
	cb.state = Closed
	slog.Info("Circuit reset to closed state")
}
