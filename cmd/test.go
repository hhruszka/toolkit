package cmd

import (
	"bptvnftester/reports"
	"bptvnftester/testengine"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"slices"
	"strings"
	"time"
)

func NewCmdTest(ctx context.Context, appName, appVersion string) *cobra.Command {
	var (
		failedonly bool
		detailed   bool
		format     string
		timeout    time.Duration
		tests      []string
	)

	cmd := &cobra.Command{
		Use:   "test [flags] [test ids]",
		Short: "Execute all or selected test cases. If run without arguments, all test cases are executed.",
		Long: `Execute all or selected test cases. If run without arguments, all test cases are executed.
If run with root account, test cases are executed for all users.
If run with non-root account, test cases are executed for the current user.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if timeout < 0 {
				_ = cmd.Usage()
				return errors.New("wrong value for timeout option! timeout cannot be a negative value, aborting")
			}

			format = strings.ToLower(format)
			supportedFormats := []string{"text", "txt", "json", "csv", "xlsx", "xls", "excel"}
			if !slices.Contains(supportedFormats, format) {
				_ = cmd.Usage()
				return fmt.Errorf("%s is not a valid report format for the output option (-o or --output), aborting", format)
			}

			// the '--detailed-report' flag can be set only with text report format.
			if detailed && (format == "xlsx" || format == "xlx" || format == "json") {
				return fmt.Errorf("--detailed-report flag can only be used with text report format")
			}

			// the '--failed-only' flag can be set only with text report format.
			if failedonly && (format == "xlsx" || format == "xlx" || format == "json") {
				return fmt.Errorf("--failed-only flag can only be used with text report format")
			}

			// tests is a CLI positional option used to pass test cases that are to be executed.
			// Test cases can be passed comma-separated or space-separated.
			for _, arg := range args {
				for _, test := range strings.Split(arg, ",") {
					if strings.TrimSpace(test) != "" {
						tests = append(tests, test)
					}
				}
			}

			// verification of test cases passed through cli.
			for _, testId := range tests {
				if !testengine.IsTestCase(strings.ToUpper(testId)) {
					return fmt.Errorf("The %[2]s is not a valid test case id. Usee '%[1]s list' to list valid test cases. Aborting!", os.Args[0], testId)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, tests, failedonly, detailed, format, timeout, appName, appVersion)
		},
	}

	cmd.Flags().BoolVarP(&failedonly, "failed-only", "", false, "create a report with failed only test cases")
	cmd.Flags().BoolVarP(&detailed, "detailed-report", "", false, "create a report with execution details for ticketing")
	cmd.Flags().StringVarP(&format, "output", "o", "xlsx", "report format: text, json, or excel (xlsx, xls)")
	cmd.Flags().DurationVarP(&timeout, "timeout", "t", time.Second*15, "timeout in seconds")
	return cmd
}

// run executes a set of tests for users based on context, generates a report, and returns any errors encountered.
func run(ctx context.Context, tests []string, failedOnly bool, detailed bool, format string, timeout time.Duration, appName, appVersion string) error {
	var testsResults []*testengine.AccountTestResults

	users := testengine.GetUsers()

	if os.Getuid() == 0 && len(users) > 0 {
		testsResults = testengine.RunAccountTests(ctx, tests, users, timeout)
	} else {
		shell, ok := users[UserName]
		if !ok {
			return fmt.Errorf("user %s not found", UserName)
		}
		// `shell -i -c bash` forces shell to run in the interactive mode, which ensures that the shell environment looks in the similar way when
		// a threat actor get initial footprint on the system (shell). This behaviour of a shell was confirmed with the shell man pages. Usage of `-i`
		// is unconditional - does not conflict with `-c`. The shell in the interactive mode sources rc files (.bashrc, .zshrc etc.) that contains
		// user account customizations e.g. additional env. variables or aliases.
		// Execution of `bash` as a command/argument of `-c` makes `bash` to inherit the environment created by `shell -i` since `bash` is launched
		// as a child process.
		//testsResults = append(testsResults, testengine.NewAccountTestResults(UserName, testengine.RunTests(tests, shell, []string{"-i", "-c", "sh"}, timeout)))
		testsResults = append(testsResults, testengine.NewAccountTestResults(UserName, testengine.RunTests(ctx, tests, shell, nil, timeout)))
	}

	if failedOnly {
		filterFailedOnly(testsResults)
	}

	return reports.GenReport(testsResults, HostName, format, detailed, appVersion)
}

// filterFailedOnly removes successful test results from the provided slice of user test results.
func filterFailedOnly(userTestsResults []*testengine.AccountTestResults) {
	for _, testsResults := range userTestsResults {
		for testId, test := range testsResults.ExecTestStatuses {
			if test.Result {
				delete(testsResults.ExecTestStatuses, testId)
			}
		}
	}
}
