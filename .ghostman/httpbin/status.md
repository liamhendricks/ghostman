---
ghostman_version: 1
method: GET
base_url: "{{env:base_url}}"
path: /status/200
---

# Status Code Request

Returns the status code specified in the path. Swap `200` for any code
(e.g. `/status/404`, `/status/500`) to test error handling.

ghostman exits non-zero on 4xx/5xx, making this useful in CI assertions.

**Usage:**

```sh
# Succeeds (exit 0)
ghostman request --env local httpbin/status

# Fails (exit 1) — useful for testing error paths
# Edit path to /status/404 first
ghostman request --env local httpbin/status || echo "request failed as expected"
```
