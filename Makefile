.PHONY: build install clean test lint verify

VERSION ?= dev

build:
	go build -ldflags "-s -w -X github.com/TeplrGuy/multiterm/cmd.version=$(VERSION)" -o multiterm .

install: build
	mkdir -p $(HOME)/.local/bin
	cp multiterm $(HOME)/.local/bin/multiterm
	@echo "Installed to $(HOME)/.local/bin/multiterm"

uninstall:
	rm -f $(HOME)/.local/bin/multiterm

clean:
	rm -f multiterm

test:
	go test ./... -count=1 -timeout 60s

lint:
	go vet ./...

verify: build
	./multiterm --version
	./multiterm --help
	./multiterm list
	@echo "✓ All checks passed"

release-dry:
	goreleaser release --snapshot --clean
