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

fmt:
	gofmt -w -s *.go

cover:
	 go test -v -covermode=count -coverprofile=coverage.out
	 go tool cover -html=coverage.out
