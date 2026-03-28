---
ghostman_version: 1
method: GET
base_url: "{{env:base_url}}"
path: /get
---

# GET Request

Sends a GET request to httpbin and returns the request details as JSON —
including your IP, headers, and query parameters.

**Usage:**

```sh
ghostman request --env local httpbin/get
```

**Pipe examples:**

```sh
# Extract your IP address
ghostman request --env local httpbin/get | ghostman get .origin

# Assert the response came from httpbin
ghostman request --env local httpbin/get | ghostman assert '.url == "https://httpbin.org/get"'
```
