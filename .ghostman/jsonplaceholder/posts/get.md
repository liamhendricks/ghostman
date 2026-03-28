---
ghostman_version: 1
method: GET
base_url: "{{env:jp_base_url}}"
path: /posts/{{col:post_id}}
---

# Get Post

Fetches a single post by ID. The ID comes from `{{col:post_id}}` — set it
with `set_collection` after a list request, or update `vars.yaml` directly.

**Usage:**

```sh
ghostman request --env local jsonplaceholder/posts/get
```

**Set post_id from a list response:**

```sh
ghostman request --env local jsonplaceholder/posts/list \
  | ghostman --col jsonplaceholder set_collection .0.id

# Now fetch that post
ghostman request --env local jsonplaceholder/posts/get
```
