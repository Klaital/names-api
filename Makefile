GOOS := linux
VERSION := 1.0.0

.PHONY: build image push

build: bin/names-web


bin/names-web: web/*.go pkg/config/*.go pkg/people/*.go
    go build -o bin/names-web ./cmd/web

image:
    docker build -t klaital/names-web$(VERSION) .

push: image
    docker push klaital/names-web$(VERSION)

run: image
    docker run --rm klaital/names-web$(VERSION)
