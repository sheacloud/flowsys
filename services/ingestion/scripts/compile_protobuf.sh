#!/bin/bash

# meant to be run from /ingestion directory
protoc -I=./internal/schema/ ./internal/schema/flow.proto --go_out=./internal/schema/ --go_opt=paths=source_relative
