package main

import (
	"os"

	lint "github.com/duh-rpc/duhrpc-lint"
)

func main() {
	os.Exit(lint.RunCmd(os.Stdin, os.Stdout, os.Stderr, os.Args[1:]))
}
