#!/bin/bash -xe

if [ -f /etc/nginx/nginx.template.conf ]; then
  cp -f /etc/nginx/nginx.template.conf /etc/nginx/nginx.conf.tmp
  if [ -n "$ENVSUBST_ENABLED" ]; then
    echo "notice: doing envsubst ..."
    IFS=$'\n'
    for K in $(printenv | cut -d= -f1); do
      echo "$0 - envsubst for $K ..."
      cat /etc/nginx/nginx.conf.tmp | envsubst '${'$K'}' | sponge /etc/nginx/nginx.conf.tmp
    done
  else
    echo "notice: ENVSUBST_ENABLED env var is missing, skip envsubst"
  fi
  mv -f /etc/nginx/nginx.conf.tmp /etc/nginx/nginx.conf
fi

/prometheus-nginx-socket-exporter &

exec "$@"

