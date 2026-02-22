.PHONY: test fmt vet

test:
	go test -v -race ./...

fmt:
	gofmt -w .

vet:
	go vet ./...
