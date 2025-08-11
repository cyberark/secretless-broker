FROM ruby:3-slim-bookworm
RUN gem update --system
RUN gem install bundler jekyll
RUN mkdir /src
COPY ./Gemfile** /src/
WORKDIR /src
RUN bundle install
