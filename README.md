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

## Development

The generated API commands live in `remnawave-api.go`. The OpenAPI snapshot is
kept in `remnawave-api.json`.

```sh
go test ./...
go build ./...
```

Releases are built by GitHub Actions from the current Remnawave OpenAPI spec at
`https://cdn.docs.rw/docs/openapi.json`.
