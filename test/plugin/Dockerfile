FROM secretless-dev:latest

WORKDIR /secretless/test/plugin/

RUN mkdir -p /usr/local/lib/secretless && \
    go build -buildmode=plugin \
             -o /usr/local/lib/secretless/example-plugin.so \
                 ./example/cmd

# Do not remove this - we are intentionally trying to exercise
# limited user functionality
USER secretless
