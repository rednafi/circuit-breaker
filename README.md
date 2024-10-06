# circuit-breaker

A minimal circuit breaker implementation in Go that manages service request flow by
switching between `closed`, `open`, and `half-open` states.

## How it works

-   **closed**: The circuit is closed, and requests are allowed through. If requests
    repeatedly fail and hit the `FailureThreshold`, the circuit moves to the `open` state.

-   **open**: The circuit is open, and all requests are blocked for the duration specified
    by `RecoveryTime`.

-   **half-open**: After the `RecoveryTime` elapses, the circuit allows a limited number of
    requests, specified by `HalfOpenMaxRequests`, to test if the service has recovered.
    Based on the success or failure of these requests, the circuit either moves back to
    `closed` or `open`.

## Installation

Install with:

```sh
go get github.com/rednafi/circuit-breaker/cb
```

## Usage

### Defining configuration

Define the configuration using the `CircuitBreakerConfig` struct:

```go
import "time"

type CircuitBreakerConfig struct {
    FailureThreshold    int           // Maximum number of failures before the circuit trips to open
    RecoveryTime        time.Duration // Duration to wait before transitioning to half-open state
    HalfOpenMaxRequests int           // Number of requests to allow in half-open state to check if service has recovered
}
```

### Creating a circuit breaker instance

Initialize the circuit breaker using values from the config struct:

```go
c := CircuitBreakerConfig{
    FailureThreshold:    3,                 // After 3 consecutive failures, the circuit will open
    RecoveryTime:        5 * time.Second,   // Wait 5 seconds before testing in half-open state
    HalfOpenMaxRequests: 2,                 // Allow 2 requests in half-open state before deciding to fully open or close
}

cb := cb.NewCircuitBreaker(c.FailureThreshold, c.RecoveryTime, c.HalfOpenMaxRequests)
```

### Making a service call

To make a request through the circuit breaker, use the `Call` method:

```go
result, err := cb.Call(unreliableService)
```

## Example usage

### Create a circuit breaker

```go
c := CircuitBreakerConfig{
    FailureThreshold:    3,               // Maximum 3 failures before circuit trips
    RecoveryTime:        5 * time.Second, // Time to wait before transitioning to half-open state
    HalfOpenMaxRequests: 2,               // Allow 2 test requests in half-open state
}

cb := cb.NewCircuitBreaker(c.FailureThreshold, c.RecoveryTime, c.HalfOpenMaxRequests)
```

### Handling successful requests in closed state

```go
successFn := func() (any, error) { return 42, nil }
result, err := cb.Call(successFn)
```

### Transition from open to half-open

```go
time.Sleep(2 * time.Second)
result, err := cb.Call(successFn)
```

### Simulating a timeout

```go
timeoutFn := func() (any, error) {
    time.Sleep(3 * time.Second)
    return nil, errors.New("timeout")
}
result, err := cb.Call(timeoutFn)
```

## Complete example

```go
// main.go

package main

import (
	"github.com/rednafi/circuit-breaker/cb"
	"time"
)

func main() {
	fn := func() (any, error) {
		return "Hello, World!", nil
	}
	cb := cb.NewCircuitBreaker(3, 5*time.Second, 2) // Initialize with 3 failure threshold, 5 sec recovery, 2 half-open requests

	result, err := cb.Call(fn)
	if err != nil {
		panic(err)
	}
	println(result.(string))
}
```

Running this with `go run main.go` prints:

```txt
2024/10/06 14:22:40 INFO Making a request state=closed
2024/10/06 14:22:40 INFO Request succeeded in closed state. Circuit remains closed.
2024/10/06 14:22:40 INFO Circuit reset to closed state.
Hello, World!
$ go run main.go
```
