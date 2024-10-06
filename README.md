# Circuit breaker

A minimal circuit breaker implementation in Go that manages service request flow by
switching between `closed`, `open`, and `half-open` states.

## How it works

- **Closed**: The circuit is closed and requests are allowed through. If requests repeatedly
  fail and hit the `FailureThreshold`, the circuit moves to the `open` state.

- **Open**: The circuit is open, and all requests are blocked for the duration specified by
  `RecoveryTime`.

- **Half-open**: After the `RecoveryTime` elapses, the circuit allows a limited number of
requests, specified by `HalfOpenMaxRequests`, to test if the service has recovered. Based on
the success or failure of these requests, the circuit either moves back to `closed` or
`open`.

## Usage

### Defining configuration

Define the configuration using the `CircuitBreakerConfig` struct:

```go
type CircuitBreakerConfig struct {
    FailureThreshold    int
    RecoveryTime        time.Duration
    HalfOpenMaxRequests int
}
```

### Creating a circuit breaker instance

Initialize the circuit breaker using the values from the config struct:

```go
c := CircuitBreakerConfig{
    FailureThreshold:    3,
    RecoveryTime:        5 * time.Second,
    HalfOpenMaxRequests: 2,
}

cb := NewCircuitBreaker(c.FailureThreshold, c.RecoveryTime, c.HalfOpenMaxRequests)
```

### Making a service call

To make a request through the circuit breaker, use the `Call` method:

```go
result, err := cb.Call(unreliableService)
```

### Example usage

#### Create a circuit breaker

```go
c := CircuitBreakerConfig{
    FailureThreshold:    3,
    RecoveryTime:        5 * time.Second,
    HalfOpenMaxRequests: 2,
}

cb := NewCircuitBreaker(c.FailureThreshold, c.RecoveryTime, c.HalfOpenMaxRequests)
```

#### Handling successful requests in closed state

```go
successFn := func() (any, error) { return 42, nil }
result, err := cb.Call(successFn)
```

#### Transition from open to half-open

```go
time.Sleep(2 * time.Second)
result, err := cb.Call(successFn)
```

#### Simulating a timeout

```go
timeoutFn := func() (any, error) {
    time.Sleep(3 * time.Second)
    return nil, errors.New("timeout")
}
result, err := cb.Call(timeoutFn)
```
