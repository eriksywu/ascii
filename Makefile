local: build-local

build-local:
	@go build -o ascii -v cmd/main.go

build-docker:
	@docker build -t ascii:latest .

unit-test:
	@go test -v ./...
