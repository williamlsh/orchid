#!/usr/bin/env bash

# Sign up api
curl "localhost:8080/api/signup" \
    -i \
    -vv \
    -X POST \
    -H "Content-Type:application/json" \
    -d '{"email": "overseas-stu@outlook.com"}'
## Response:
# {"code":0,"message":"Success"}
## Mail content: http://localhost/m/callback?token=wxcjwAuZCkCj&operation=login&state=overseatu

# -------------------------------------------------------------------------------------------------------------

## Sign in api
curl "localhost:8080/api/signin" \
    -i \
    -vv \
    -X POST \
    -H "Content-Type:application/json" \
    -d '{"alias":"william","code":"NKHcknTxImyE","operation":"register"}'

## Reponse:
# {"code":0,"message":"Success","data":{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImI4YTAzMzZmLWFjNTctNDJjZC1iZDc1LTMwZDQyOGQ2ZTg5OCIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDYzMTE5NCwidXNlcl9pZCI6MTE3NjUzOTcyfQ.biKYq25lclAwxwpWAsH0I1DfSRie2x-GjQatJfqly3w","refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTEyMzUwOTQsInJlZnJlc2hfdXVpZCI6ImI4YTAzMzZmLWFjNTctNDJjZC1iZDc1LTMwZDQyOGQ2ZTg5OCsrMTE3NjUzOTcyIiwidXNlcl9pZCI6MTE3NjUzOTcyfQ.0pd3DvCixrfzCdEyBmZOCI95tCHBjf-IhuJDVdEvvYg"}}

# -------------------------------------------------------------------------------------------------------------

# Token refresh api
curl "localhost:8080/api/token/refresh" \
    -i \
    -vv \
    -X POST \
    -H "Content-Type:application/json" \
    -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTEzMDQ0ODUsInJlZnJlc2hfdXVpZCI6ImMwNTRmOWZhLWE5MjgtNDAwNC04YTVmLTBmOWRkMjUxMDc4YysrMTE3NjUzOTcyIiwidXNlcl9pZCI6MTE3NjUzOTcyfQ.zl-Qc5t3QozaWWRDzWX1V4TadTb399ShPd4L8P5SDek"}'

## Response:
# {
#   "code": 0,
#   "message": "Success",
#   "data": {
#     "access_token": "xxx",
#     "refresh_token": "xxx"
#   }
# }

# -------------------------------------------------------------------------------------------------------------

# Sign out api
curl "localhost:8080/api/signout" \
    -i \
    -vv \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6IjY2ZTFjY2Y3LWFmZGYtNDFjYy05YmM2LTcyOTg3MWU5OWM4MiIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDcwMTg2MCwidXNlcl9pZCI6MTE3NjUzOTcyfQ.DvwkQYHADpWhJV9iQze_vbqxt_MQmTriL9cPIdYgHSA"

## Response:
# {"code":0,"message":"Success"}
curl "localhost:8080/api/signout" \
    -i \
    -vv \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImM2NzU3YTA1LTg0ZjUtNDE4ZC1hZDVkLTFmOWNlZGZmY2YxZSIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDcwMDYxNSwidXNlcl9pZCI6MTE3NjUzOTcyfQ.1FRv1W3aQoNXLOQj3IWmyvCtffLuPhS49Y9V2ixsoFo"

# -------------------------------------------------------------------------------------------------------------

# Deregister api
curl "localhost:8080/api/deregister" \
    -i \
    -vv \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImRmMWE1YTE0LTZhMzktNDVhNy1iYWEyLTIyMjQ4ZDRlMzNjNSIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDcwMTc2MiwidXNlcl9pZCI6MTE3NjUzOTcyfQ.F1tF_EgCtee9HRGUp1dVM3gH3SgWBnGZHlTIR1mnV0Y"

## Response:
# {"code":0,"message":"Success"}
