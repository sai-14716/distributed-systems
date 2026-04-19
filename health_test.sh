#!/usr/bin/env bash

URL="http://localhost:8088/health"
INTERVAL=0.5

while true; do
  TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")

  RESPONSE=$(curl -s -w "\n%{http_code}" "$URL")

  BODY=$(echo "$RESPONSE" | sed '$d')
  STATUS=$(echo "$RESPONSE" | tail -n1)

  echo "[$TIMESTAMP] Status: $STATUS"
  echo "$BODY"
  echo "----------------------------------"

  sleep "$INTERVAL"
done