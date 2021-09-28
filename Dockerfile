FROM alpine:latest

ARG INSIGHTS_CLI_VERSION ${INSIGHTS_CLI_VERSION}
ENV RELEASE_URL "https://github.com/FairwindsOps/insights-cli/releases/download/v${INSIGHTS_CLI_VERSION}/insights-cli_${INSIGHTS_CLI_VERSION}_linux_amd64.tar.gz"

WORKDIR /

RUN mkdir -p /build && cd build && \
    wget $RELEASE_URL && \
    tar -xvf insights-cli_${INSIGHTS_CLI_VERSION}_linux_amd64.tar.gz insights-cli -C /usr/local/bin/ && \
    chmod a+x /usr/local/bin/insights-cli && \
    rm -rf /build

CMD /usr/local/bin/insights-cli
