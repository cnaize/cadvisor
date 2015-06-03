#!/bin/bash

set -e
set -x

godep go build -a github.com/cnaize/cadvisor

docker build -t google/cadvisor:canary .
