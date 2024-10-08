package cb

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedStateSuccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(3, 5*time.Second, 3, 2*time.Second) // Updated argument order

	successFn := func() (any, error) {
		return 42, nil
	}

	result, err := cb.Call(successFn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if val, ok := result.(int); !ok || val != 42 {
		t.Fatalf("expected result 42, got %v", result)
	}

	if cb.state != Closed {
		t.Fatalf("expected state closed, got %s", cb.state)
	}
}

func TestCircuitBreaker_ClosedStateFailure(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(2, 5*time.Second, 3, 2*time.Second) // Updated argument order

	failFn := func() (any, error) {
		return nil, errors.New("failure")
	}

	// First failure
	_, err := cb.Call(failFn)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	// Second failure should trigger state change to open
	_, err = cb.Call(failFn)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if cb.state != Open {
		t.Fatalf("expected state open, got %s", cb.state)
	}
}

func TestCircuitBreaker_OpenToHalfOpen(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(1, 1*time.Second, 2, 2*time.Second) // Updated argument order

	failFn := func() (any, error) {
		return nil, errors.New("failure")
	}

	// Circuit is closed, so calling should allow it first
	_, _ = cb.Call(failFn)

	// After the first failure, the circuit should transition to open
	_, err := cb.Call(failFn)
	if err == nil || err.Error() != "circuit open, request blocked" {
		t.Fatalf("expected error 'circuit open, request blocked', got %v", err)
	}

	// Simulate time passing to trigger recovery and transition to half-open
	time.Sleep(2 * time.Second)

	// After recovery, the next call should transition to half-open, no error expected
	_, err = cb.Call(failFn)
	if err != nil {
		t.Fatalf("expected no error during transition to half-open, got %v", err)
	}

	if cb.state != HalfOpen {
		t.Fatalf("expected state half-open, got %s", cb.state)
	}

	// Now simulate a successful request
	successFn := func() (any, error) {
		return 42, nil
	}

	result, err := cb.Call(successFn)
	if err != nil {
		t.Fatalf("expected no error on successful request, got %v", err)
	}

	if val, ok := result.(int); !ok || val != 42 {
		t.Fatalf("expected result 42, got %v", result)
	}

	// Ensure the state is still half-open after the first success
	if cb.state != HalfOpen {
		t.Fatalf("expected state half-open after first success, got %s", cb.state)
	}

	// Another successful request should transition the breaker to closed
	result, err = cb.Call(successFn)
	if err != nil {
		t.Fatalf("expected no error on second successful request, got %v", err)
	}

	if val, ok := result.(int); !ok || val != 42 {
		t.Fatalf("expected result 42, got %v", result)
	}

	// Ensure the state is now closed after enough successes
	if cb.state != Closed {
		t.Fatalf("expected state closed after two successful requests, got %s", cb.state)
	}
}

func TestCircuitBreaker_HalfOpenStateFailure(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(1, 1*time.Second, 2, 2*time.Second) // Updated argument order

	cb.state = HalfOpen

	failFn := func() (any, error) {
		return nil, errors.New("failure")
	}

	// Call in half-open state should transition back to open on failure
	_, err := cb.Call(failFn)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if cb.state != Open {
		t.Fatalf("expected state open, got %s", cb.state)
	}
}

func TestCircuitBreaker_OpenToHalfOpenSuccess(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(1, 1*time.Second, 1, 2*time.Second) // Updated argument order

	// Simulate a failure to trigger transition to open
	failFn := func() (any, error) {
		return nil, errors.New("failure")
	}

	_, err := cb.Call(failFn)
	if err == nil {
		t.Fatalf("expected error during failure, got nil")
	}

	// Ensure the breaker is now in the Open state
	if cb.state != Open {
		t.Fatalf("expected state open after failure, got %s", cb.state)
	}

	// Simulate time passing to trigger recovery and transition to half-open
	time.Sleep(2 * time.Second)

	// First successful request should transition to half-open
	successFn := func() (any, error) {
		return 42, nil
	}

	_, err = cb.Call(successFn)
	if err != nil {
		t.Fatalf("expected no error during transition to half-open, got %v", err)
	}

	// Check that the state is now half-open
	if cb.state != HalfOpen {
		t.Fatalf("expected state half-open, got %s", cb.state)
	}

	// Another successful request should transition to closed
	_, err = cb.Call(successFn)
	if err != nil {
		t.Fatalf("expected no error during successful request in half-open state, got %v", err)
	}

	// Ensure the breaker is now closed after enough successful requests
	if cb.state != Closed {
		t.Fatalf("expected state closed, got %s", cb.state)
	}
}

func TestCircuitBreaker_RequestTimeout(t *testing.T) {
	t.Parallel()

	cb := NewCircuitBreaker(2, 1*time.Second, 3, 2*time.Second) // Updated argument order

	// Simulate a service call that hangs (takes longer than the timeout)
	timeoutFn := func() (any, error) {
		time.Sleep(3 * time.Second)
		return nil, errors.New("timeout")
	}

	_, err := cb.Call(timeoutFn)
	if err == nil || err.Error() != "request timed out" {
		t.Fatalf("expected timeout error, got %v", err)
	}
}
