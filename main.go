package main

import (
	"github.com/danielgtaylor/openapi-cli-generator/cli"
)

func main() {
	cli.Init(&cli.Config{
		AppName:   "remnawave-cli",
		EnvPrefix: "REMNAWAVE_CLI",
		Version:   "1.0.0",
	})

	remnawaveApiRegister(false)

	cli.Root.Execute()
}
