FROM ruby:2.3.7
RUN gem install bundler jekyll
RUN mkdir /src
COPY ./Gemfile** /src/
WORKDIR /src
RUN bundle install
