package main

// Only used for example

import (
	"errors"
	"github.com/rednafi/circuit-breaker/cb"
	"log"
	"log/slog"
	"time"
)

func unreliableService() (any, error) {
	if time.Now().Unix()%2 == 0 {
		return 0, errors.New("service failed")
	}
	return 42, nil
}

func main() {
	cb := cb.NewCircuitBreaker(
		2,             // Failure threshold
		2*time.Second, // Recovery time
		2,             // Half-open max requests
		2*time.Second, // Timeout
	)

	for i := 0; i < 5; i++ {
		result, err := cb.Call(unreliableService)
		if err != nil {
			slog.Error("Service request failed", "error", err)
		} else {
			slog.Info("Service request succeeded", "result", result)
		}

		time.Sleep(1 * time.Second)
		log.Println("-----------------------------------------------")
	}
}
