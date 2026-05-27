# remnawave-cli

Generated command-line client for the Remnawave API.

## Install

```sh
curl -fsSL https://raw.githubusercontent.com/intezya/remnawave-cli/main/install.sh | sh
```

By default the installer downloads the latest GitHub Release and installs
`remnawave-cli` into `/usr/local/bin` when possible, falling back to
`~/.local/bin`. You can override the target with `INSTALL_DIR`:

```sh
curl -fsSL https://raw.githubusercontent.com/intezya/remnawave-cli/main/install.sh | INSTALL_DIR="$HOME/bin" sh
```

## Usage

```sh
remnawave-cli --help
remnawave-cli --server https://panel.example.com authcontroller-login '{"username":"admin","password":"password"}'
```

## Agent-Optimized Output

The default output remains full JSON. For agent workflows, use compact formats
that reduce token-heavy API responses:

```sh
remnawave-cli userscontroller-getallusers -o agent --fields uuid,username,status --limit 25
remnawave-cli userscontroller-getallusers --output ndjson --fields uuid,username,status
```

Available response output formats:

- `json`: full pretty JSON, default
- `yaml`: full YAML
- `agent`: compact summary plus a tab-separated table
- `ndjson`: one JSON object per row

Useful agent flags:

- `--fields uuid,username,status`: keep only selected fields in `agent`/`ndjson`
- `--limit 25`: cap rows shown by `agent`; use `0` to disable truncation
- `--save-full /tmp/remnawave-response.json`: save the full JSON response before compacting
- ``--query 'users[?status==`ACTIVE`]'``: apply a JMESPath projection before formatting

## Development

The generated API commands live in `remnawave-api.go`. The OpenAPI snapshot is
kept in `remnawave-api.json`.

```sh
go test ./...
go build ./...
```

Releases are built by GitHub Actions from the current Remnawave OpenAPI spec at
`https://cdn.docs.rw/docs/openapi.json`.
