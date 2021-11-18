#!/bin/sh

echo "=== /starting prometheus-nginx-socket-exporter ==="
/prometheus-nginx-socket-exporter &

echo "=== starting cmd: $* ==="
exec "$@"

