FROM ruby:3-slim-bullseye
RUN gem update --system
RUN gem install bundler jekyll
RUN mkdir /src
COPY ./Gemfile** /src/
WORKDIR /src
RUN bundle install
