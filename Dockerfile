FROM alpine

RUN apk update && apk add --no-cache mysql-client curl

COPY bin/names-web /names-web

EXPOSE 8080
USER nobody:nobody
ENTRYPOINT ["/names-web"]

