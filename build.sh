#!/usr/bin/env bash

# Not strictly necessary, but a little more contained than using 'go install'

mkdir -p bin

for dir in ./cmd/*; do
    go build -o "bin/${dir##*/}" "$dir"
done

