#!/bin/bash

# meant to be run from /ingestion directory
protoc -I=./ ./flow.proto --go_out=./ --go_opt=paths=source_relative
