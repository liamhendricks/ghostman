---
ghostman_version: 1
method: GET
base_url: "{{env:base_url}}"
path: /secure
cert: "{{env:client_cert}}"
key: "{{env:client_key}}"
---

# mTLS Request

A request that requires a client certificate.
