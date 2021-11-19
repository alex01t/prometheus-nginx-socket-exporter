FROM golang:1.17 as build
WORKDIR /app
COPY . .
RUN go mod vendor
RUN go build

FROM openresty/openresty:latest
RUN apt-get -qq update && apt-get install -qq net-tools procps curl vim && rm -rf /var/lib/apt/lists/*
COPY --from=build /app/prometheus-nginx-socket-exporter /prometheus-nginx-socket-exporter
COPY nginx/lua/monitor.lua /etc/nginx/lua/
COPY nginx/conf.d/monitor.conf nginx/conf.d/default.conf /etc/nginx/conf.d/

COPY ./entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/openresty", "-g", "daemon off;"]
