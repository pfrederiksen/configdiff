FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY configdiff /usr/bin/configdiff

ENTRYPOINT ["/usr/bin/configdiff"]
