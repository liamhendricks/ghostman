---
steps:
  - ghostman request --env local httpbin/post
  - ghostman assert '.json.tool == "ghostman"'
---

# Post and Assert

Sends a POST request and asserts the body was echoed back correctly.
Exits non-zero if the assertion fails — useful in CI pipelines.

Run: `ghostman chain post-and-assert`
