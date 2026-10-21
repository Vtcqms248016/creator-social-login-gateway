#!/bin/sh
set -eu

: "${CAPTCHA_TOKEN:?set CAPTCHA_TOKEN to a browser-issued Turnstile token}"

curl --fail-with-body \
  --request POST \
  --header 'Content-Type: application/json' \
  --data "{\"provider\":\"google\",\"captcha_token\":\"${CAPTCHA_TOKEN}\",\"creator_id\":\"cr_17\",\"asset_id\":\"asset_9\",\"next\":\"asset_delivery\"}" \
  http://localhost:8080/login/start
