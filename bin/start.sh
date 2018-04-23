#!/bin/bash

set -m

trap 'kill -TERM $(jobs -p)' TERM INT

line_server /var/data &

# Templates
ep /etc/nginx/sites-available/default

nginx
