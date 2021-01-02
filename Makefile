VERSION=v0.4.0

build:
	go build -o ./bin/site-deploy  -ldflags="-s -w -X main.version=$(VERSION)" -trimpath ./cmd/site-deploy

test:
	test -d /tmp/example.com || mkdir -p /tmp/example.com
	go test ./cmd/site-deploy
