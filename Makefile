.PHONY: build test lint fmt clean

build:
	go build -o kyaraben ./cmd/kyaraben

test:
	go test ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .
	goimports -w -local github.com/fnune/kyaraben .

clean:
	rm -f kyaraben
