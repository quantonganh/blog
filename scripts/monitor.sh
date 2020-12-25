#!/bin/sh

script_name=$(basename "$0")
domain=$1
status=$(/usr/bin/curl -s -w "%{http_code}" -I -X GET "$domain" -o /dev/null)
if [ "$status" -ne 200 ]; then
  /usr/bin/osascript -e "display notification \"$domain is down\" with title \"$script_name\""
fi
