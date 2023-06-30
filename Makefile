
.PHONY: build
build:
	docker-compose build

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	@go test -v \
		-count=1 \
		-timeout=5m \
		./...

.PHONY: run
run:
	docker-compose up -d

.PHONY: stop
stop:
	docker-compose down