#!/bin/bash

DEFAULT_SCHEME="http"
read -p "Choose the default scheme [http/https] (default is http): " scheme
SCHEME=${scheme:-$DEFAULT_SCHEME}

SWAGGER_BINARY=$(cat .found-swagger-path)/dist/swagger

if [ "$1" = "dev" ]; then
    $SWAGGER_BINARY generate client -f swagger-dev.json -m kbmodel -c kbclient --default-scheme=$SCHEME

    # If dev extensions are generated, delete the non-dev client file if it exists
    if [ -f "./kbclient/kill_bill_client.go" ]; then
        rm "./kbclient/kill_bill_client.go"
    fi
else
    $SWAGGER_BINARY generate client -f kbswagger.yaml -m kbmodel -c kbclient --default-scheme=$SCHEME

    # If non-dev extensions are generated, delete the dev client file if it exists
    if [ -f "./kbclient/kill_bill_dev_client.go" ]; then
        rm "./kbclient/kill_bill_dev_client.go"
    fi
fi
