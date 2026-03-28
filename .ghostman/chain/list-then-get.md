---
steps:
  - ghostman request --env local jsonplaceholder/posts/list
  - ghostman --col jsonplaceholder set_collection .0.id
  - ghostman request --env local jsonplaceholder/posts/get
---

# List Then Get

Fetches a list of posts, stores the first post's ID in the collection vars,
then fetches that specific post. Demonstrates chaining list → detail flows.

1. GET /posts?userId=1 → list of posts
2. Extract `.0.id` → store as `post_id` in jsonplaceholder/vars.yaml
3. GET /posts/{{col:post_id}} → single post detail

Run: `ghostman chain list-then-get`
