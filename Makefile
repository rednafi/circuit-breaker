.PHONY: $(MAKECMDGOALS)

lint:
	go fmt ./...

test:
	go test -v ./...

