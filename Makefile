VERSION=v0.1.0

build:
	go build -o ./bin/site-deploy  -ldflags="-s -w -X main.version=$(VERSION)" -trimpath ./cmd/site-deploy

test:
	go test ./cmd/site-deploy
