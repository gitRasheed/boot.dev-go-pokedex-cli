package main

import (
	"os"
	"github.com/gitRasheed/boot.dev-go-pokedex-cli/pkg/cli"
)

func main() {
	cli.Run(os.Stdin, os.Stdout)
}