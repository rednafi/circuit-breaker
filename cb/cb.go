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

type circuitBreaker struct {
	FailureThreshold     int
	FailureCount         int
	RecoveryTime         time.Duration
	State                string
	LastFailureTime      time.Time
	HalfOpenSuccessCount int
	HalfOpenMaxRequests  int
	mu                   sync.Mutex
}

func NewCircuitBreaker(
	failureThreshold int, recoveryTime time.Duration,
	halfOpenMaxRequests int,
) *circuitBreaker {
	return &circuitBreaker{
		FailureThreshold:    failureThreshold,
		RecoveryTime:        recoveryTime,
		State:               Closed,
		HalfOpenMaxRequests: halfOpenMaxRequests,
	}
}

func (cb *circuitBreaker) Call(fn func() (any, error)) (any, error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	slog.Info("Making a request", "state", cb.State)
	switch cb.State {
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

func (cb *circuitBreaker) handleClosedState(
	fn func() (any, error),
) (any, error) {
	result, err := cb.runWithTimeout(fn)
	if err != nil {
		slog.Warn(
			"Request failed in closed state. Incrementing failure count.",
		)
		cb.recordFailure()
		return nil, err
	}
	slog.Info("Request succeeded in closed state. Circuit remains closed.")
	cb.reset() // Reset after a successful request
	return result, nil
}

func (cb *circuitBreaker) handleOpenState() (any, error) {
	if time.Since(cb.LastFailureTime) > cb.RecoveryTime {
		slog.Info(
			"Recovery period expired. Transitioning to half-open state.",
		)
		cb.State = HalfOpen
		cb.FailureCount = 0 // Reset failure count in half-open state
		cb.HalfOpenSuccessCount = 0
		return nil, nil
	}
	slog.Warn("Circuit is still open. Blocking requests.")
	return nil, errors.New("circuit is open. Blocking request.")
}

func (cb *circuitBreaker) handleHalfOpenState(
	fn func() (any, error),
) (any, error) {
	result, err := cb.runWithTimeout(fn)
	if err != nil {
		slog.Error(
			"Request failed in half-open state. Circuit transitioning back.",
		)
		cb.State = Open
		cb.LastFailureTime = time.Now()
		return nil, err
	}

	cb.HalfOpenSuccessCount++
	slog.Info(
		"Request succeeded in half-open state.",
		"successCount", cb.HalfOpenSuccessCount,
		"maxRequests", cb.HalfOpenMaxRequests,
	)

	// If enough successful requests are made, transition to closed state
	if cb.HalfOpenSuccessCount >= cb.HalfOpenMaxRequests {
		slog.Info(
			"Enough successful requests in half-open state. Transitioning.",
		)
		cb.reset()
	}

	return result, nil
}

func (cb *circuitBreaker) runWithTimeout(
	fn func() (any, error),
) (any, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(), 2*time.Second,
	)
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

func (cb *circuitBreaker) recordFailure() {
	cb.FailureCount++
	cb.LastFailureTime = time.Now()

	if cb.FailureCount >= cb.FailureThreshold {
		slog.Error(
			"Failure threshold reached. Circuit transitioning to open.",
			"failureCount", cb.FailureCount,
			"threshold", cb.FailureThreshold,
		)
		cb.State = Open
	} else {
		slog.Warn(
			"Failure recorded",
			"failureCount", cb.FailureCount,
			"threshold", cb.FailureThreshold,
		)
	}
}

func (cb *circuitBreaker) reset() {
	cb.FailureCount = 0
	cb.State = Closed
	slog.Info("Circuit reset to closed state.")
}
