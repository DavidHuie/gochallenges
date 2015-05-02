default: test

test:
	go test ./...

deps:
	go get -u github.com/golang/lint/golint golang.org/x/tools/cmd/vet

vet:
	go vet ./...

lint:
	golint ./...
