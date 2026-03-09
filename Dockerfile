FROM alpine:3.23

RUN apk update && apk -U upgrade --no-cache
USER nobody
# The insights-cli binary will have been built by goreleaser.
COPY insights-cli /
ENTRYPOINT ["/insights-cli"]
