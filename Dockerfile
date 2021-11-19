FROM golang:1.17 as build
WORKDIR /app
COPY . .
RUN go mod vendor
RUN go build

FROM openresty/openresty:latest
RUN apt-get -qq update && apt-get install -qq procps vim
COPY --from=build /app/prometheus-nginx-socket-exporter /prometheus-nginx-socket-exporter
COPY monitor.lua /etc/nginx/lua/
COPY monitor.conf default.conf /etc/nginx/conf.d/

COPY ./entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/openresty", "-g", "daemon off;"]
