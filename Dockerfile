FROM golang:1.17 as build
WORKDIR /app
COPY . .
RUN go mod vendor
RUN go build

FROM openresty/openresty:latest
RUN apt-get -qq update && apt-get install -qq procps vim
COPY --from=build /app/prometheus-nginx-socket-exporter /prometheus-nginx-socket-exporter
COPY monitor.lua /etc/nginx/lua/
COPY default.conf /etc/nginx/conf.d/
