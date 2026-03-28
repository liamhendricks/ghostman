---
ghostman_version: 1
method: GET
base_url: "{{env:jp_base_url}}"
path: /posts
query_params:
  userId: "{{col:user_id}}"
---

# List Posts

Fetches all posts for a given user. The `userId` query param comes from
`{{col:user_id}}` in the collection vars — change it with `set_collection`.

**Usage:**

```sh
ghostman request --env local jsonplaceholder/posts/list
```

**Pipe examples:**

```sh
# Count posts returned
ghostman request --env local jsonplaceholder/posts/list \
  | ghostman get .

# Store the first post ID for use in other requests
ghostman request --env local jsonplaceholder/posts/list \
  | ghostman --col jsonplaceholder set_collection .0.id
```
