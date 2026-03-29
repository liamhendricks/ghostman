# ghostman

A terminal-native API client written in Go. Requests, collections, and environments in plain markdown files. They live in your repo, render as documentation, and compose via Unix pipes! Ghostman also allows you to define `ghostman chain` files for maintaining request test flows. Want to write your own scripts? Ghostman provides the ability for users to write their own Go files to execute arbitrary logic and pipe it into the chain with `ghostman script`. And more!

```sh
ghostman request --env staging auth/login \
  | ghostman assert '.status == "ok"' \
  | ghostman set_env --env-file .ghostman/env/staging.yaml token .data.token
```

## Install

```sh
go install github.com/liamhendricks/ghostman/cmd/ghostman@latest
```

Requires Go 1.24+.

## Ghostman concepts

### Requests

A request is a markdown file with YAML frontmatter. It lives in a collection directory in any project under `.ghostman/` or globally inside `~/.ghostman/`.

```
.ghostman/
  auth/
    login.md
    refresh.md
  users/
    list.md
    get.md
  env/
    local.yaml
    staging.yaml
  chain/
    test-auth-flow
```

**`auth/login.md`:**

```markdown
---
ghostman_version: 1
method: POST
base_url: "{{env:base_url}}"
path: /auth/login
headers:
  Content-Type: application/json
---

# Login

Authenticates a user and returns a user object.

```json
{
  "username": "user",
  "password": "pass"
}
```

### Environments

Environments are YAML files in `env/` inside any collection root.

```yaml
# .ghostman/env/staging.yaml
base_url: https://api.staging.example.com
username: testuser
password: hunter2
```

Reference them in requests with `{{env:key}}`. Pass `--env <name>` to select one at runtime.

### Collections

A collection is any directory containing request files. The `--col` flag targets a specific collection for `set_collection`. Collections can have a `vars.yaml` file for shared runtime state.

```yaml
# .ghostman/auth/vars.yaml
token: eyJhbGci
```

Reference collection vars in requests with `{{col:key}}`.

### Config

`~/.ghostmanrc` (YAML) lets you register additional collection roots to reference globally:

```yaml
collections:
  - /home/you/shared-requests
```

---

## Commands

### `ghostman request`

Execute a request file and write the response body to stdout.

```sh
ghostman request <collection/name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--env <name>` | Load `env/<name>.yaml` for variable substitution |
| `--dry-run` | Print the substituted request without sending it |
| `--insecure` | Skip TLS certificate verification |

**Examples:**

```sh
# Basic request
ghostman request auth/login

# With environment
ghostman request --env staging auth/login

# Dry run — inspect substituted values
ghostman request --env staging auth/login --dry-run

# Pipe to another command
ghostman request --env staging users/list | jq '.data[]'
```

**Exit codes:** `0` on 2xx, `1` on 4xx/5xx or network error. The response body is always written to stdout even on error, so it's available downstream.

**mTLS:** Add `cert` and `key` fields to the request frontmatter:

```yaml
---
ghostman_version: 1
method: GET
base_url: "{{env:base_url}}"
path: /secure
cert: "{{env:client_cert}}"
key: "{{env:client_key}}"
---
```

---

### `ghostman list`

List all available requests across all collection roots.

```sh
ghostman list
```

```
auth/login
auth/refresh
users/get
users/list
```

---

### `ghostman get`

Extract a value from piped JSON using a dot-notation path.

```sh
... | ghostman get <path>
```

Uses [gjson](https://github.com/tidwall/gjson) path syntax. Leading dots are stripped automatically.

```sh
ghostman request --env staging auth/login | ghostman get .data.token
# → eyJhbGci...

ghostman request --env staging users/list | ghostman get .data.0.id
# → 42
```

**Note:** `ghostman get` is a terminal step — it outputs the extracted value with a newline and does **not** pass through the original JSON. Use `assert` or `set_env` if you need to keep the JSON flowing.

---

### `ghostman assert`

Assert a condition on piped JSON. Passes JSON through on success; exits non-zero on failure.

```sh
... | ghostman assert '<expression>'
```

**Supported operators:** `==` and `!=`

```sh
ghostman request --env staging auth/login \
  | ghostman assert '.status == "ok"' \
  | ghostman get .data.token
```

Exits `1` with a descriptive error if the assertion fails. Useful as a CI gate:

```sh
ghostman request --env prod healthcheck/ping \
  | ghostman assert '.status == "healthy"' \
  || exit 1
```

---

### `ghostman set_env`

Extract a value from piped JSON and write it to an environment YAML file. Passes JSON through.

```sh
... | ghostman set_env --env-file <path> <key> <json-path>
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--env-file <path>` | Path to the env YAML file to update (required) |

```sh
ghostman request --env staging auth/login \
  | ghostman set_env --env-file .ghostman/env/staging.yaml token .data.token \
  | ghostman assert '.status == "ok"'
```

This updates `staging.yaml` in place:

```yaml
base_url: https://api.staging.example.com
token: eyJhbGci...   # ← written by set_env
```

---

### `ghostman set_collection`

Extract a value from piped JSON and write it to a collection's `vars.yaml`. Passes JSON through.

```sh
... | ghostman --col <collection> set_collection <json-path>
```

The key name is derived from the last segment of the path: `.data.token` → key `token`.

```sh
ghostman request --env staging auth/login \
  | ghostman --col auth set_collection .data.token
# writes token: eyJhbGci... to .ghostman/auth/vars.yaml
```

The stored value is then available as `{{col:token}}` in any request that uses the `auth` collection.

---

### `ghostman chain`

Execute a named chain file — a sequence of steps with stdout piped to stdin between them. Aborts on the first non-zero exit.

```sh
ghostman chain <name>
```

Chain files live at `.ghostman/chain/<name>.md`:

```markdown
---
steps:
  - ghostman request --env staging auth/login
  - ghostman assert '.status == "ok"'
  - ghostman --col auth set_collection .data.token
  - ghostman request --env staging users/list
---

# Auth and List Users

Logs in, validates the response, stores the token, then fetches users.
```

Steps can be any `ghostman` subcommand. Each step receives the previous step's stdout as its stdin.

**Exit codes** are preserved — if a step exits `1`, `ghostman chain` exits `1` with the same code.

---

### `ghostman script`

Run a user-written Go file as a pipe step. Receives JSON on stdin, writes JSON to stdout.

```sh
... | ghostman script <file.go>
```

The script must be in (or reachable from) a directory with a `go.mod` that includes ghostman as a dependency — typically your project root. It's executed via `go run`.

**Writing a script:**

```go
package main

import "github.com/liamhendricks/ghostman/pkg/script"

func main() {
    script.RunFunc(func(d script.Data) (script.Data, error) {
        name := d.Get("name")
        return d.Set("greeting", "Hello, "+name+"!")
    })
}
```

Or implement the `Handler` interface for more structure:

```go
package main

import (
    "strings"
    "github.com/liamhendricks/ghostman/pkg/script"
)

type UppercaseToken struct{}

func (u *UppercaseToken) Handle(d script.Data) (script.Data, error) {
    token := d.Get("data.token")
    return d.Set("data.token", strings.ToUpper(token))
}

func main() {
    script.Run(&UppercaseToken{})
}
```

**`pkg/script` API:**

| Symbol | Description |
|--------|-------------|
| `type Handler interface` | `Handle(Data) (Data, error)` — implement this or use `RunFunc` |
| `type Data struct` | Wraps raw JSON bytes |
| `Data.Get(path string) string` | Extract a value by gjson path |
| `Data.Set(path, value string) (Data, error)` | Return new Data with value set at path |
| `Data.Raw() []byte` | Return the underlying JSON bytes |
| `Run(h Handler)` | Read stdin → call handler → write stdout |
| `RunFunc(fn func(Data) (Data, error))` | Convenience wrapper for function literals |
| `NewData(raw []byte) Data` | Construct Data directly (useful in tests) |

**In a chain file:**

```yaml
steps:
  - ghostman request --env staging auth/login
  - ghostman assert '.status == "ok"'
  - ghostman script .ghostman/scripts/enrich-token.go
  - ghostman --col auth set_collection .data.token
```

---

## Pipe model

Every command follows the same contract:

- Reads JSON from **stdin**
- Writes JSON to **stdout**
- Writes diagnostics (status, duration, errors) to **stderr**
- Exits non-zero on failure

This means any combination composes:

```sh
# Validate, transform, store, and extract — all in one pipeline
ghostman request --env prod auth/login \
  | ghostman assert '.status == "ok"' \
  | ghostman script .ghostman/scripts/normalize.go \
  | ghostman set_env --env-file .ghostman/env/prod.yaml token .data.token \
  | ghostman get .data.user_id
```

Commands that are not terminal steps (`assert`, `set_env`, `set_collection`, `script`) pass the original JSON through unchanged so the pipeline keeps flowing.

---

## Variable substitution

Two namespaces are available in any request field (URL, path, headers, body, cert/key paths):

| Syntax | Source |
|--------|--------|
| `{{env:key}}` | Active environment YAML file |
| `{{col:key}}` | Collection `vars.yaml` |

Unknown namespaces and missing keys are hard errors — ghostman will not silently send a request with unresolved placeholders.

**Path parameters** use `:param` syntax and are resolved from the environment:

```yaml
path: /users/:user_id
```

---

## Request file format

```markdown
---
ghostman_version: 1
method: POST
base_url: "{{env:base_url}}"
path: /users/:user_id/posts
headers:
  Authorization: Bearer {{col:token}}
  Content-Type: application/json
query:
  include_deleted: "false"
cert: "{{env:client_cert}}"   # optional — mTLS client cert path
key: "{{env:client_key}}"     # optional — mTLS client key path
---

# Create Post

Optional markdown body — renders as documentation.

```json
{
  "title": "title",
  "body": "Hello"
}
```

The code block language determines the `Content-Type` if not set explicitly:

| Language | Content-Type |
|----------|-------------|
| `json` | `application/json` |
| `xml` | `application/xml` |
| `form` | `application/x-www-form-urlencoded` |
| `graphql` | `application/json` |
| `body` | no Content-Type set (escape hatch) |

---

## Project layout

```
pkg/
  chain/       Parse and execute chain files
  collection/  Discover requests, load/write vars.yaml
  config/      Parse ~/.ghostmanrc
  env/         Load env YAML, variable substitution, update env files
  executor/    Execute HTTP requests
  parser/      Parse request markdown → RequestSpec
  pipe/        JSON utilities (ReadJSON, Extract, Assert)
  script/      Handler interface, Data type, Run/RunFunc
cmd/ghostman/  CLI commands (thin wrappers over pkg/)
.ghostman/     Example requests, chains, and scripts
```

`pkg/` has zero Cobra/Viper imports — it's a pure library usable outside the CLI.
