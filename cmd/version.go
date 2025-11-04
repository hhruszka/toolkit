package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

func NewCmdVersion(appName, appVersion, buildTime, gitCommit string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, _ []string) {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), appName, "version: "+appVersion, "build time: "+buildTime, "git commit: "+gitCommit)
		},
	}
	return cmd
}
