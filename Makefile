GOOS := linux
VERSION := latest

.PHONY: build image push run

build: bin/names-web


bin/names-web: cmd/web/*.go pkg/config/*.go pkg/people/*.go
	go build -o bin/names-web ./cmd/web

image: build
	docker build -t klaital/names-web:$(VERSION) .

push: image
	docker push klaital/names-web:$(VERSION)

run: image
	docker-compose up
