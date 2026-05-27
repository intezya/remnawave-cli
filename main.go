package main

import (
	"github.com/danielgtaylor/openapi-cli-generator/cli"
)

var version = "dev"

func main() {
	cli.Init(&cli.Config{
		AppName:   "remnawave-cli",
		EnvPrefix: "REMNAWAVE_CLI",
		Version:   version,
	})

	configureAgentOutput()
	remnawaveApiRegister(false)

	cli.Root.Execute()
}
