local: build-local

build-local:
	@go build -ldflags="-X 'main.StorePath=tempstore'" -o ascii -v cmd/main.go

build:
	@docker build -t ascii:latest .

unit-test:
	@go test -v ./...

docker-run:
	@echo "making volume on host"
	@mkdir -p ${CURDIR}/asciistore
	@docker run -d -p 8000:8000 --mount type=bind,source=${CURDIR}/asciistore,target=/asciistore ascii:latest

service: build docker-run

test: unit-test
	@./run_tests.sh
