FROM alpine

RUN apk update && apk add postgresql-client curl postgresql-dev
# Clean APK cache
RUN rm -rf /var/cache/apk/*

COPY bin/names-web /names-web

EXPOSE 8080
USER nobody:nobody
ENTRYPOINT ["/names-web"]
