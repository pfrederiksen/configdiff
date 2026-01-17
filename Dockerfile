FROM --platform=$BUILDPLATFORM alpine:latest AS certs
RUN apk add --no-cache ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY configdiff /usr/bin/configdiff

ENTRYPOINT ["/usr/bin/configdiff"]
