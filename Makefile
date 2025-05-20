BIN = bin


.PHONY: test
test:
	go test ./...

.PHONY: coverage
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

clean_coverage:
	rm -f coverage.out coverage.html

build:
	go build -o $(BIN)/gw

clean_build:
	rm -rf $(BIN)

clean: clean_coverage clean_build
