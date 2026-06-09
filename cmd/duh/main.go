package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/duh-rpc/duh-cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	os.Exit(duh.RunCmd(ctx, os.Stdout, os.Args[1:]))
}
