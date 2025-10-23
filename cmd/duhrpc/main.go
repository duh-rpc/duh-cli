package main

import (
	"os"

	"github.com/duh-rpc/duhrpc"
)

func main() {
	os.Exit(duhrpc.RunCmd(os.Stdout, os.Args[1:]))
}
