package main

import (
	"context"
	"os"

	"github.com/mszostok/gh-discover/cmd"
	"github.com/mszostok/gh-discover/pkg/xsignal"
)

func main() {
	rootCmd := cmd.NewRoot()
	ctx := xsignal.WithStopContext(context.Background())

	if err := rootCmd.ExecuteContext(ctx); err != nil {
		// error is already handled by `cobra`, we don't want to log it here as we will duplicate the message.
		// If needed, based on error type we can exit with different codes.
		os.Exit(1)
	}
}
