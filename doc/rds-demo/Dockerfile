FROM ubuntu:16.04

RUN apt-get update && \
    apt-get install -y build-essential \
                       curl \
                       groff \
                       jq \
                       less \
                       postgresql-client \
                       postgresql-contrib \
                       python \
                       python-dev \
                       python-pip \
                       software-properties-common \
                       vim

RUN apt-add-repository ppa:brightbox/ruby-ng && \
    apt-get update && \
    apt-get install -y ruby2.4 ruby2.4-dev

RUN pip install awscli

RUN gem install -N conjur-cli -v '~> 6.0'
