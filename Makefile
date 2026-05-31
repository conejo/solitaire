.PHONY: build test run-web run-tui clean

BINARY := solitaire

build:
	go build -o $(BINARY) ./cmd/solitaire

test:
	go test ./...

run-web: build
	./$(BINARY) web

run-tui: build
	./$(BINARY) tui

clean:
	rm -f $(BINARY)
