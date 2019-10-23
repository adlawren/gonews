#!/usr/bin/env bash

if [ "$GONEWS_ENV" = "DEV" ]
then
    /usr/bin/env bash
else
    gonews
fi
