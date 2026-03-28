---
steps:
  - ghostman request --env local httpbin/get
  - ghostman --col httpbin set_collection .origin
  - ghostman script .ghostman/scripts/example.go
  - ghostman get .origin
---

# Test Flow

Smoke test chain: sends a GET request, stores the origin IP in the httpbin
collection vars, then extracts and prints it.

Run: `ghostman chain test-flow`
