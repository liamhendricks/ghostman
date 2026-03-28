---
ghostman_version: 1
method: POST
base_url: "{{env:base_url}}"
path: /upload
---

# Upload

Multiple code blocks but body tag disambiguates.

```json
{"example": "this is documentation, not body"}
```

```body
{"actual": "request body"}
```
