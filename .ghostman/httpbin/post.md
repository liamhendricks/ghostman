---
ghostman_version: 1
method: POST
base_url: "{{env:base_url}}"
path: /post
headers:
  Content-Type: application/json
---

# POST Request

Sends a JSON body to httpbin. The response echoes back the parsed body
under `.json`, headers, and request metadata.

**Usage:**

```sh
ghostman request --env local httpbin/post
```

**Pipe examples:**

```sh
# Assert the body was received correctly
ghostman request --env local httpbin/post \
  | ghostman assert '.json.tool == "ghostman"'

# Extract just the echoed body
ghostman request --env local httpbin/post | ghostman get .json
```

```json
{
  "tool": "ghostman",
  "description": "terminal-native API client"
}
```
