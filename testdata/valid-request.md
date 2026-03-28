---
ghostman_version: 1
method: POST
base_url: "{{env:base_url}}"
path: /users/:id
headers:
  Content-Type: application/json
  Authorization: "Bearer {{env:token}}"
query_params:
  page: "1"
  filter: "{{col:filter}}"
---

# Create User

Creates a new user in the system.

```json
{
  "name": "{{env:name}}",
  "role": "admin"
}
```
