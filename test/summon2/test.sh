#!/bin/bash -e

# TODO: I am not sure why `go test -v .`` doesn't run both of these
go test -v summon2_cmd_test.go
go test -v summon2_run_test.go
