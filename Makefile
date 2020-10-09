.PHONY: test lint

test:
	go test -v -cover

lint:
	go fmt
	go vet ./...
	golint ./...
