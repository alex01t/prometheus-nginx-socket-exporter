#!/bin/sh

/prometheus-nginx-socket-exporter &

exec "$@"

