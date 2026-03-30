VERSION ?= dev

.PHONY: build build-all test lint release clean

build:
	CGO_ENABLED=0 go build -ldflags="-X main.version=$(VERSION)" -o bin/envguard .

build-all:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o dist/envguard-darwin-amd64 .
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o dist/envguard-darwin-arm64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o dist/envguard-linux-amd64 .
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-X main.version=$(VERSION)" -o dist/envguard-linux-arm64 .
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-X main.version=$(VERSION)" -o dist/envguard-windows-amd64.exe .

test:
	go test ./... -v -race

lint:
	golangci-lint run

release:
	goreleaser release --snapshot --clean

clean:
	rm -rf bin/ dist/
