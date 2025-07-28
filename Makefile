build-server:
  go build -o bin/trivia-server ./cmd/server

build-client:
  go build -o bin/trivia-client ./cmd/client

run-server:
  go run ./cmd/server

run-client:
  go run ./cmd/client

test:
  go test ./...
