FROM ubuntu:16.04

RUN apt update && \
    apt install -y \
    nginx \
    xz-utils \
    vim \
    gettext-base

COPY bin/* /usr/bin/

COPY conf/nginx.conf /etc/nginx/
COPY conf/default.conf /etc/nginx/sites-available/default.tmpl


CMD ["start.sh"]
