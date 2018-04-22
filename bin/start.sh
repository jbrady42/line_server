#!/bin/bash

set -m

trap 'kill -TERM $(jobs -p)' TERM INT

line_server /var/data &

envsubst < /etc/nginx/sites-available/default.tmpl > /etc/nginx/sites-available/default

nginx
