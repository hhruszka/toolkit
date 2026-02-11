package cmd

import (
	"bptvnftester/log"
	"context"
	"fmt"
	"os"
	"os/user"
	"strconv"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// App global variables
var (
	HostName string
	UserName string
	UserID   int
)

func Init() {
	var err error

	UserID = os.Getuid()
	UserName = strconv.Itoa(UserID)
	userInfo, err := user.Current()
	if err != nil {
		fmt.Println("Failed to get current user info:", err)
		os.Exit(1)
	}
	UserName = userInfo.Username

	HostName, _ = os.Hostname()
	//fmt.Println("Host:", HostName, "Uid:", UserID, "User:", UserName)
}

var appZapLogger *zap.Logger

func setupLogger(logLevel string, appName, appVersion string) (*zap.Logger, error) {
	var err error

	appLogFile := fmt.Sprintf("%s.%s", appName, "log")
	if _, err := os.Stat(appLogFile); err == nil {
		if err := os.Remove(appLogFile); err != nil {
			return nil, fmt.Errorf("cannot remove %s log file due to : %w", appLogFile, err)
		}
	}

	appLogger, err := log.InitLogger(logLevel, appLogFile)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Replace the global Zap logger with our configured instance.
	// This makes zap.L() and zap.S() available throughout the application.
	zap.ReplaceGlobals(appLogger)

	zap.L().Info("Logger initialized via Cobra flags", zap.String("configured_level", logLevel))
	zap.L().Info("Application starting up...", zap.String("version", appVersion))
	return appLogger, nil
}

func NewRootCmd(ctx context.Context, appName, appVersion, buildTime, gitCommit string) *cobra.Command {
	var logFlg bool
	var debugFlg bool

	Init()

	cmd := &cobra.Command{
		Use:           appName + " [flags] [test ids]",
		Short:         appVersion + " is a command line application that executes baseline penetration test cases on hosts (VMs, k8s nodes etc.)",
		Long:          ``,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			var err error
			var logLevel string

			// these flags are mutually exclusive, so only one of them will be set
			if logFlg || debugFlg {
				if logFlg {
					logLevel = "info"
				}

				if debugFlg {
					logLevel = "debug"
				}

				if appZapLogger, err = setupLogger(logLevel, appName, appVersion); err != nil {
					return fmt.Errorf("failed to initialize logger: %w", err)
				}
			}

			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if appZapLogger != nil {
				appZapLogger.Sync()
				log.CloseLogger()
			}
			return nil
		},
	}

	// Disable automatic printing of usage when an error occurs
	cmd.SilenceUsage = true

	// support for '--'
	cmd.Flags().SetInterspersed(false)

	cmd.PersistentFlags().BoolVar(&logFlg, "log", false, "enable logging")
	cmd.PersistentFlags().BoolVar(&debugFlg, "debug", false, "enable debug logging")
	_ = cmd.PersistentFlags().MarkHidden("debug")
	cmd.MarkFlagsMutuallyExclusive("log", "debug")

	cmd.AddCommand(NewCmdTest(ctx, appName, appVersion), NewCmdList(), NewCmdVersion(appName, appVersion, buildTime, gitCommit))

	return cmd
}
