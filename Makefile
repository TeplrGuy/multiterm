.PHONY: build install clean test lint

VERSION ?= dev

build:
	go build -ldflags "-s -w -X github.com/gilbertappiah/multiterm/cmd.version=$(VERSION)" -o multiterm .

install: build
	cp multiterm /usr/local/bin/multiterm

uninstall:
	rm -f /usr/local/bin/multiterm

clean:
	rm -f multiterm

test:
	go test ./...

lint:
	go vet ./...

release-dry:
	goreleaser release --snapshot --clean
