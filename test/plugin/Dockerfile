FROM secretless-dev:latest

WORKDIR /secretless/test/plugin/

COPY . .

RUN mkdir -p /usr/local/lib/secretless && \
    go build -buildmode=plugin \
             -o /usr/local/lib/secretless/example-plugin.so \
                 ./example/

# Do not remove this - we are intentionally trying to exercise
# limited user functionality
USER secretless
