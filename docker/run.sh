#!/usr/bin/env bash

if [ "$GONEWS_ENV" = "DEV" ]
then
    go build
    ./gonews &
    /usr/bin/env bash
else
    gonews
fi
