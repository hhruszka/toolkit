package main

import (
	"bptvnftester/cmd"
	"context"
	"fmt"
	"os"
	"os/signal"
)

var (
	AppName    = "bptvnftester"
	AppVersion = "dev"     // Set by -ldflags
	BuildTime  = "unknown" // Set by -ldflags
	GitCommit  = "unknown" // Set by -ldflags
	GitBranch  = "unknown" // Set by -ldflags
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	rootCmd := cmd.NewRootCmd(ctx, AppName, AppVersion, BuildTime, GitCommit)
	if err := rootCmd.Execute(); err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
}
