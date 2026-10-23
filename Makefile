BINARY := astock

.PHONY: build fetch serve schedule vet clean

build:
	go build -o $(BINARY) ./cmd/astock

fetch:
	go run ./cmd/astock fetch

serve:
	go run ./cmd/astock serve

schedule:
	go run ./cmd/astock schedule -every 1h

vet:
	go vet ./...
	go build ./...

clean:
	rm -f $(BINARY) $(BINARY).exe
