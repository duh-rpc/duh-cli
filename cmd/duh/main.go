package main

import (
	"os"

	"github.com/duh-rpc/duh-cli"
)

func main() {
	os.Exit(duh.RunCmd(os.Stdout, os.Args[1:]))
}
