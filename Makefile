BINARY := logtimeline
PKG := ./...

.PHONY: build test benchmark fmt lint clean

build:
	go build -o bin/$(BINARY) .

test:
	go test $(PKG)

benchmark:
	go test -bench=. -benchmem $(PKG)

fmt:
	gofmt -w $$(find . -name '*.go')

lint:
	go vet $(PKG)

clean:
	rm -rf bin
