.PHONY: build test bench lint run clean download-blocklist

build:
	go build -o bin/sinkhole .

test:
	go test -v -race -count=1 ./...

bench:
	go test -bench=. -benchmem ./...

lint:
	go vet ./...

run: build
	./bin/sinkhole --blocklist=blocklist.txt --verbose

clean:
	rm -rf bin/

download-blocklist:
	curl -sL https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts -o blocklist.txt
