#!/bin/sh
# Entrypoint for the quickproxy container.
#
#   docker run ... ghcr.io/OWNER/REPO:latest                 # serve
#   docker run -it ... ghcr.io/OWNER/REPO:latest login -manual
#
# config.json is created on first start with an address reachable from outside
# the container; every omitted field falls back to the built-in default.
set -eu

CONFIG="${QW_CONFIG:-/data/config.json}"

if [ ! -f "$CONFIG" ]; then
    mkdir -p "$(dirname "$CONFIG")"
    printf '{\n  "addr": "0.0.0.0:8080",\n  "database": "quickwork.db"\n}\n' > "$CONFIG"
    chmod 0600 "$CONFIG"
fi

# Respect a -config passed explicitly on the command line.
for arg in "$@"; do
    case "$arg" in
        -config|--config|-config=*|--config=*)
            exec /usr/local/bin/quickproxy "$@"
            ;;
    esac
done

exec /usr/local/bin/quickproxy "$@" -config "$CONFIG"
