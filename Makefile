.PHONY: all tidy test cover

# Compile
all:
	go build .

# Clean up go.mod
tidy:
	go mod tidy

# Run tests
test:
	go test

# Test coverage
cover:
	go test -coverprofile cover.prof
	go tool cover -html=cover.prof
