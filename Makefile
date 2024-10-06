.PHONY: $(MAKECMDGOALS)

lint:
	@go fmt ./...
	@prettier --write .

lint-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Go code is not formatted. Run 'go fmt' on your code."; \
		exit 1; \
	fi
	@prettier --check . > /dev/null || (echo "Prettier check failed. Run 'prettier --write .' to fix." && exit 1)


test:
	@go test -v ./...
