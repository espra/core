#! /usr/bin/env bash

source ./environ/common.sh

counter=0

function step() {
    ((counter++)) || true
    printf "\n\033[34;1m➡  Running $@  \033[90m[step $counter] [running ${SECONDS}s]\033[0m\n\n"
    eval $@
}

step go mod download
step go test -coverprofile coverage.txt -v ./...
