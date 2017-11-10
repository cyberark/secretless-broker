#!/bin/bash -ex

godep restore
go test ./cmd/secretless
