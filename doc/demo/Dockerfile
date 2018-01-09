FROM ubuntu:16.04

RUN apt-get update && apt-get install -y vim jq curl postgresql-client python python-dev software-properties-common build-essential postgresql-contrib groff less software-properties-common
RUN curl -Ls -o /tmp/get-pip.py https://bootstrap.pypa.io/get-pip.py && python2 /tmp/get-pip.py
RUN pip install awscli

RUN apt-add-repository ppa:brightbox/ruby-ng && apt-get update
RUN apt-get install -y ruby2.4 ruby2.4-dev
RUN gem install -N conjur-cli -v '~> 6.0'
