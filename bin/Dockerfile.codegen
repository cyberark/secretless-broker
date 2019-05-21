FROM golang:1.12.5

WORKDIR /

ENV GO111MODULE=on

RUN git clone --depth 1 \
              -b fix-runners \
              https://github.com/conjurinc/code-generator.git && \
   cd /code-generator && \
   ./build_generators && \
   cp -r dist/* /usr/local/bin

WORKDIR /secretless
