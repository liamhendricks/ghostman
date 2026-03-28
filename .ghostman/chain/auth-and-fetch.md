---
steps:
  - ghostman request --env local httpbin/post
  - ghostman --col httpbin set_collection .json.tool
  - ghostman request --env local httpbin/anything
---

# Auth and Fetch

Demonstrates the full collection state pattern:

1. POST to get a "token" (httpbin echoes the body, we use `.json.tool`)
2. Store the token in the httpbin collection vars via `set_collection`
3. Make an authenticated request that uses `{{col:token}}` from vars

This mirrors a real login → use token flow:
  - Replace step 1 with your auth endpoint
  - Replace `.json.tool` with the path to the token in the response (e.g. `.data.token`)
  - Replace step 3 with the protected endpoint

Run: `ghostman chain auth-and-fetch`
