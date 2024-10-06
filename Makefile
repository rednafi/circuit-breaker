.PHONY: $(MAKECMDGOALS)

lint:
	go fmt ./...
	prettier --write .

test:
	go test -v ./...
