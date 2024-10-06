# circuit-breaker

A minimal circuit breaker implementation in Go that manages service request flow by
switching between `closed`, `open`, and `half-open` states.

## How it works

-   **closed**: Requests are allowed through. If requests repeatedly fail and hit the
    `FailureThreshold`, the circuit moves to `open`.
-   **open**: Requests are blocked for the duration specified by `RecoveryTime`.
-   **half-open**: After `RecoveryTime`, a limited number of requests
    (`HalfOpenMaxRequests`) are allowed to test if the service has recovered. Based on the
    outcome, the circuit either moves back to `closed` or `open`.

## Installation

```sh
go get github.com/rednafi/circuit-breaker/cb
```

## Usage

This example demonstrates both a successful request and a failed request using the circuit
breaker.

```go
package main

import (
    "errors"
    "fmt"
    "github.com/rednafi/circuit-breaker/cb"
    "time"
)

func main() {
    // Initialize circuit breaker with:
    // - 3 failures threshold
    // - 5 seconds recovery time
    // - 2 requests allowed in half-open state
    // - 2 seconds request timeout
    circuitBreaker := cb.NewCircuitBreaker(3, 5*time.Second, 2, 2*time.Second)

    // Simulating a successful service request
    successFn := func() (any, error) {
        return "Success!", nil
    }

    result, err := circuitBreaker.Call(successFn)
    if err != nil {
        fmt.Printf("Request failed: %v\n", err)
    } else {
        fmt.Printf("Request succeeded: %v\n", result)
    }

    // Simulating a failed service request
    failFn := func() (any, error) {
        return nil, errors.New("Service failure")
    }

    result, err = circuitBreaker.Call(failFn)
    if err != nil {
        fmt.Printf("Request failed: %v\n", err)
    } else {
        fmt.Printf("Request succeeded: %v\n", result)
    }
}
```

1. The first request simulates a **successful call** by returning `"Success!"` with no
   error.
2. The second request simulates a **failed call** by returning an error `"Service failure"`.

You can modify the `failFn` or `successFn` to see how the circuit breaker behaves when
transitioning between states like `closed`, `open`, and `half-open`.
