package main

import (
	"os"

	lint "github.com/duh-rpc/duhrpc-lint"
)

func main() {
	os.Exit(lint.RunCmd(os.Stdout, os.Args[1:]))
}
