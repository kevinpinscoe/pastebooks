#!/usr/bin/env bash

USER="userid"
PC="password to test login with"

# Login and capture cookie
curl -sk -c /tmp/pb.cookies \
  -H 'Content-Type: application/json' \
  -X POST \
  --data '{"email":"$USER","passcode":"$PC"}' \
  https://kevin.kevininscoe.com/pastebooks/api/login

# Call /api/me with that cookie
curl -sk -b /tmp/pb.cookies \
  https://kevin.kevininscoe.com/pastebooks/api/me
