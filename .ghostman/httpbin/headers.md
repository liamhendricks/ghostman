---
ghostman_version: 1
method: GET
base_url: "{{env:base_url}}"
path: /headers
headers:
  X-Request-Id: ghostman-demo
  Accept: application/json
---

# Inspect Request Headers

Returns the headers that ghostman sent with the request.
Useful for verifying that custom headers are being forwarded correctly.

**Usage:**

```sh
ghostman request --env local httpbin/headers
```

**Pipe examples:**

```sh
# Extract just the headers object
ghostman request --env local httpbin/headers | ghostman get .headers

# Assert custom header was sent
ghostman request --env local httpbin/headers \
  | ghostman assert '.headers.X-Request-Id == "ghostman-demo"'
```
