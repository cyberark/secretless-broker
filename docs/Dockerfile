FROM ruby:2.4.0
RUN gem update --system
RUN gem install bundler jekyll
RUN mkdir /src
COPY ./Gemfile** /src/
WORKDIR /src
RUN bundle install
