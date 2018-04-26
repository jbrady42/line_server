FROM golang:1.10-stretch

WORKDIR /go/src/line_server

COPY . ./

RUN make line_server


FROM ubuntu:16.04

RUN apt update && \
    apt install -y \
    curl \
    nginx \
    xz-utils \
    vim

RUN curl -sLo /usr/bin/ep \
    https://github.com/kreuzwerker/envplate/releases/download/1.0.0-RC1/ep-linux && \
    chmod +x /usr/bin/ep

RUN  mkdir -p /data/nginx/cache

COPY --from=0 /go/src/line_server/line_server /usr/bin/

COPY bin/* /usr/bin/

COPY conf/nginx.conf /etc/nginx/
COPY conf/default.conf /etc/nginx/sites-available/default


CMD ["start.sh"]
