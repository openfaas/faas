.PHONY: test cover

build:
	go build

test:
	gofmt -s -w *.go
	goimports -w *.go
	go vet ./...
	go test -v
	go test -v --race
	staticcheck ./...

	cd v2/
	gofmt -s -w *.go
	goimports -w *.go
	go vet ./...
	go test -v
	go test -v --race
	staticcheck ./...

fmt:
	gofmt -w -s *.go
	go mod tidy
	cd v2/
	gofmt -w -s *.go
	go mod tidy

cover:
	 go test -v -covermode=count -coverprofile=coverage.out
	 go tool cover -html=coverage.out
