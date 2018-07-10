FROM ruby:2.5-alpine

RUN apk add --update alpine-sdk && \
    mkdir -p /tmp/gems

WORKDIR /tmp/gems

COPY Gemfile* /tmp/gems/
RUN bundle install

WORKDIR /usr/src/app

CMD ["jekyll", "build", "--destination", "/tmp/_site"]
