#!/bin/sh
set -e

if [ "$1" = 'raindrops' ]; then

    exec "$@"
fi

exec "$@"