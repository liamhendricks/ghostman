---
ghostman_version: 1
method: POST
base_url: "{{env:jp_base_url}}"
path: /posts
headers:
  Content-Type: application/json
---

# Create Post

Creates a new post. jsonplaceholder simulates the creation and returns the
new post with an assigned `id` (always 101 in the sandbox).

**Usage:**

```sh
ghostman request --env local jsonplaceholder/posts/create
```

**Pipe examples:**

```sh
# Assert the post was created with the correct title
ghostman request --env local jsonplaceholder/posts/create \
  | ghostman assert '.title == "Hello from ghostman"'

# Store the new post ID
ghostman request --env local jsonplaceholder/posts/create \
  | ghostman --col jsonplaceholder set_collection .id
```

```json
{
  "title": "Hello from ghostman",
  "body": "Sent via terminal — no GUI required.",
  "userId": 1
}
```
