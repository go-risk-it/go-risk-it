lint:
	docker run --rm --tty\
		-v $(shell pwd):/app \
        -w /app \
        golangci/golangci-lint:v1.55.2 golangci-lint run -v
