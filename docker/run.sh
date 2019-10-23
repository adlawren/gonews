#!/usr/bin/env bash

if [ "$GONEWS_ENV" = "DEV" ]
then
    go get -u github.com/go-delve/delve/cmd/dlv
    go get github.com/golang/mock/mockgen
    /usr/bin/env bash
else
    gonews
fi
