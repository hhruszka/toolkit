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

	fmt.Println("Starting", AppName, AppVersion, "(build time: "+BuildTime+", git commit: "+GitCommit+")\n")

	rootCmd := cmd.NewRootCmd(ctx, AppName, AppVersion, BuildTime, GitCommit)
	if invokedCmd, err := rootCmd.ExecuteC(); err != nil {
		fmt.Printf("ERROR: %s\n\n", err)
		fmt.Printf("Use \"%s --help\" for more information.\n\n", AppName+" "+invokedCmd.Name())
		os.Exit(1)
	}
}
