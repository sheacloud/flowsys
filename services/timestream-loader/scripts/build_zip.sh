#!/bin/bash

GOOS=linux CGO_ENABLED=0 go build -o dynamo-loader ./cmd/loader/

zip function.zip dynamo-loader
