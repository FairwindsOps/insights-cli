FROM alpine:3.19
USER nobody
# The insights-cli binary will have been built by goreleaser.
COPY insights-cli /
ENTRYPOINT ["/insights-cli"]
