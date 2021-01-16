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
    -d '{"alias":"william","code":"sPAEePeIOikP","operation":"login"}'

## Reponse:
# {"code":0,"message":"Success","data":{"access_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImI4YTAzMzZmLWFjNTctNDJjZC1iZDc1LTMwZDQyOGQ2ZTg5OCIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDYzMTE5NCwidXNlcl9pZCI6MTE3NjUzOTcyfQ.biKYq25lclAwxwpWAsH0I1DfSRie2x-GjQatJfqly3w","refresh_token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTEyMzUwOTQsInJlZnJlc2hfdXVpZCI6ImI4YTAzMzZmLWFjNTctNDJjZC1iZDc1LTMwZDQyOGQ2ZTg5OCsrMTE3NjUzOTcyIiwidXNlcl9pZCI6MTE3NjUzOTcyfQ.0pd3DvCixrfzCdEyBmZOCI95tCHBjf-IhuJDVdEvvYg"}}

# -------------------------------------------------------------------------------------------------------------

# Token refresh api
curl "localhost:8080/api/token/refresh" \
    -i \
    -vv \
    -X POST \
    -H "Content-Type:application/json" \
    -d '{"refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2MTE0MTEzMTksInJlZnJlc2hfdXVpZCI6IjQyZTFhYzAwLTRhNTctNDk5ZC05NWQxLWM2NmFjNzkxNmVmYisrMTE3NjUzOTcyIiwidXNlcl9pZCI6MTE3NjUzOTcyfQ.exp2K0r0AepUAzMtZjmAjXjntOCZDYmLMN_0maRlNZs"}'

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
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImI2MDQ5YTEzLWFiMDEtNGNhYi05YjU5LWZhZDgxMzY3OTRkMyIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDgwNzQ0NywidXNlcl9pZCI6MTE3NjUzOTcyfQ.2cHotDy0HCOIBBvRR47OMA6gl9JbGM42f7ubO4StFzk"

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
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6Ijc2MGIxMzIzLWRiYTAtNDEwMi04NTJiLWM1OTRkOTlmNzJlNCIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDgwOTA3NywidXNlcl9pZCI6MTE3NjUzOTcyfQ.RK-BTtOwaSS9cJ3coUWS28T3JzbqxOmdiOFqyE6jHmI"

## Response:
# {"code":0,"message":"Success"}

# -------------------------------------------------------------------------------------------------------------

# Update email
curl "localhost:8080/api/account" \
    -i \
    -vv \
    -X POST \
    -H "Content-Type:application/json" \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImJiN2JlODZhLTMxM2UtNDU2Yi04MzI2LTQyNDYxMzM0ZWE4MSIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDgwOTI0MCwidXNlcl9pZCI6MTE3NjUzOTcyfQ.HS6Xuo4CzOcc7rdEhc8i6SmrWi6pWH6mgt63u1vC2WU" \
    -d '{"email": "overseas-stu@outlook.com"}'

## Response:
# {"code":0,"message":"Success"}

# -------------------------------------------------------------------------------------------------------------

# Get user profile
curl "localhost:8080/api/user/profile" \
    -i \
    -vv \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImE4ZmVlMmRmLTAzY2UtNDBjMC1iNWVlLTg1Yzk0YzU5ZjliZiIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDgxMTE3MCwidXNlcl9pZCI6MTE3NjUzOTcyfQ.BXXnQVP-5RGh8WozyI0KaUO574OvrFpFED3Byni_50E"

## Response:
# {
#   "code": 0,
#   "message": "Success",
#   "data": {
#     "username": "overseas-stu",
#     "email": "overseas-stu@outlook.com"
#   }
# }

# -------------------------------------------------------------------------------------------------------------

# Update user profile
curl "localhost:8080/api/user/profile" \
    -i \
    -vv \
    -X POST \
    -H "Content-Type:application/json" \
    -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhY2Nlc3NfdXVpZCI6ImE4ZmVlMmRmLTAzY2UtNDBjMC1iNWVlLTg1Yzk0YzU5ZjliZiIsImF1dGhvcml6ZWQiOnRydWUsImV4cCI6MTYxMDgxMTE3MCwidXNlcl9pZCI6MTE3NjUzOTcyfQ.BXXnQVP-5RGh8WozyI0KaUO574OvrFpFED3Byni_50E" \
    -d '{"username": "overseas-stu2"}'

## Response:
# {"code":0,"message":"Success"}
