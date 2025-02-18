#!/bin/bash

export REDIS_DB=0
export REDIS_PASSWORD=password
export LOG_LEVEL=debug
export PORT=8888
export LISTEN_EVENTS=true

gow -e=go,mod,html run .
