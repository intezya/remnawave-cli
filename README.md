# remnawave-cli

Generated command-line client for the Remnawave API.

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
