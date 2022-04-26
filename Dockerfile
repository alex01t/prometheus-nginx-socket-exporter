FROM golang:1.17 as build
WORKDIR /app
COPY go.mod go.sum ./
COPY main.go nginx.go nginx_status.go process.go socket.go ./
RUN go mod vendor
RUN go build

FROM openresty/openresty:1.19.9.1-5-bullseye-fat
RUN apt-get -qq update \
    && apt-get install -qq net-tools procps curl vim moreutils \
    && rm -rf /var/lib/apt/lists/*
RUN opm install ledgetech/lua-resty-http
RUN rm -rf /etc/nginx /var/log/nginx \
    && ln -sTf /usr/local/openresty/nginx/conf /etc/nginx \
    && ln -sTf /usr/local/openresty/nginx/logs /var/log/nginx
COPY --from=build /app/prometheus-nginx-socket-exporter /prometheus-nginx-socket-exporter

# enable lua monitoring
COPY ./nginx/lua/monitor.lua     /etc/nginx/monitor/
COPY ./nginx/conf.d/monitor.conf /etc/nginx/conf.d/

COPY ./entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
CMD ["/usr/bin/openresty", "-g", "daemon off;"]
