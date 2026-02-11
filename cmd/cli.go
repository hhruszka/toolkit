package cmd

import "time"

// CliOptions represents the configuration options for the command-line interface of an application or tool.
type CliOptions struct {
	AppName    string
	AppVersion string

	Timeout    time.Duration
	Debug      bool
	Format     string
	ReportFile string
	Version    bool
	Failedonly bool
	Detailed   bool
	Tests      []string
}
