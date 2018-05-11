#!/bin/bash -ex

rm -rf .env
docker-compose down -v
rm -rf run/mysql/*
