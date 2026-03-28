---
ghostman_version: 1
method: POST
base_url: "{{env:base_url}}"
path: /anything
headers:
  Content-Type: application/json
  Authorization: Bearer {{col:token}}
---

# Anything — Auth Demo

Sends a request with a Bearer token sourced from the collection vars
(`{{col:token}}`). Use `set_collection` to write the token first:

```sh
# Step 1: fetch a token (using httpbin/post as a stand-in)
ghostman request --env local httpbin/post \
  | ghostman --col httpbin set_collection .json.tool

# Step 2: send authenticated request (token now in httpbin/vars.yaml)
ghostman request --env local httpbin/anything
```

Or run the full flow as a chain:

```sh
ghostman chain --env local auth-and-fetch
```

```json
{
  "action": "demonstrate-auth",
  "token_source": "col:token"
}
```
