package cmd

import (
	"bptvnftester/reports"
	"bptvnftester/testengine"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func NewCmdTest(ctx context.Context, appName, appVersion string) *cobra.Command {
	var cliOptions = new(CliOptions)

	cliOptions.Tests = make([]string, 0)
	cliOptions.AppVersion = appVersion
	cliOptions.AppName = appName

	cmd := &cobra.Command{
		Use:   "test [flags] [test ids]",
		Short: "Execute all or selected test cases. If run without arguments, all test cases are executed.",
		Long: `Execute all or selected test cases. If run without arguments, all test cases are executed.
If run with root account, test cases are executed for all users.
If run with non-root account, test cases are executed for the current user.`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if cliOptions.Timeout < 0 {
				return errors.New("wrong value for timeout option! timeout cannot be a negative value, aborting")
			}

			if os.Geteuid() != 0 && cliOptions.Users != "all" {
				return errors.New("non-root user cannot execute tests for other users")
			}

			if err := validateFormat(cmd, cliOptions); err != nil {
				return fmt.Errorf("failed to validate format: %w", err)
			}

			// tests is a CLI positional option used to pass test cases that are to be executed.
			// Test cases can be passed comma-separated or space-separated.
			if err := validateTests(cmd, args, cliOptions); err != nil {
				return fmt.Errorf("failed to validate tests: %w", err)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(ctx, cliOptions)
		},
	}

	cmd.Flags().BoolVarP(&cliOptions.Failedonly, "failed-only", "", false, "create a report with failed only test cases")
	cmd.Flags().BoolVarP(&cliOptions.Detailed, "detailed-report", "", false, "create a report with execution details for ticketing")
	cmd.Flags().StringVarP(&cliOptions.Format, "output", "o", "xlsx", "file path or report format: text, xlsx or json.\nIf the file path is provided the extension determines the report format.")
	cmd.Flags().StringVarP(&cliOptions.Users, "users", "u", "all", "comma separated list of user accounts to test.\nThis option is only available for root users.")
	cmd.Flags().DurationVarP(&cliOptions.Timeout, "timeout", "t", time.Second*15, "timeout in seconds")
	return cmd
}

// run executes a set of tests for users based on context, generates a report, and returns any errors encountered.
func run(ctx context.Context, options *CliOptions) error {
	var testsResults []*testengine.AccountTestResults
	var cliUsers []string
	var users map[string]string = make(map[string]string)

	foundUsers := testengine.GetUsers()

	if options.Users != "all" {
		cliUsers = strings.Split(options.Users, ",")

		for _, user := range cliUsers {
			if _, ok := foundUsers[user]; !ok {
				return fmt.Errorf("user %s not found", user)
			}
			users[user] = foundUsers[user]
		}
	} else {
		users = foundUsers
	}

	if os.Getuid() == 0 && len(users) > 0 {
		testsResults = testengine.RunAccountTests(ctx, options.Tests, users, options.Timeout)
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
		testsResults = append(testsResults, testengine.NewAccountTestResults(UserName, testengine.RunTests(ctx, options.Tests, shell, nil, options.Timeout)))
	}

	if options.Failedonly {
		filterFailedOnly(testsResults)
	}

	return reports.GenReport(testsResults, HostName, options.ReportFile, options.Format, options.Detailed, options.AppVersion)
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

func validateFormat(_ *cobra.Command, cliOptions *CliOptions) error {
	var reportFile string

	supportedFormats := []string{"text", "txt", "csv", "json", "xlsx", "xls", "excel"}

	extension := filepath.Ext(cliOptions.Format)
	if extension != "" {
		reportFile = cliOptions.Format
		cliOptions.Format = extension[1:]
	}
	if extension == "" && len(reportFile) > 0 {
		return fmt.Errorf("missing extension in the provided file path %s", cliOptions.Format)
	}

	cliOptions.Format = strings.ToLower(cliOptions.Format)
	if !slices.Contains(supportedFormats, cliOptions.Format) {
		return fmt.Errorf("%s is not a valid report format for the output option (-o or --output), aborting", cliOptions.Format)
	}

	// the '--detailed-report' flag can be set only with text report format.
	if cliOptions.Detailed && (cliOptions.Format == "xlsx" || cliOptions.Format == "xls" || cliOptions.Format == "json") {
		return fmt.Errorf("--detailed-report flag can only be used with text report format")
	}

	// the '--failed-only' flag can be set only with text report format.
	if cliOptions.Failedonly && (cliOptions.Format == "xlsx" || cliOptions.Format == "xlx" || cliOptions.Format == "json") {
		return fmt.Errorf("--failed-only flag can only be used with text report format")
	}

	if reportFile != "" {
		cliOptions.ReportFile = reportFile

		reportFile = filepath.Clean(reportFile)
		stat, err := os.Stat(filepath.Dir(reportFile))
		if err == nil && stat != nil && !stat.IsDir() {
			return fmt.Errorf("the provided file path is not valid; aborting")
		}
		if err != nil {
			return fmt.Errorf("failed to access the provided file path due to: %w", err)
		}
	}
	return nil
}

func validateTests(_ *cobra.Command, args []string, cliOptions *CliOptions) error {
	for _, arg := range args {
		for _, test := range strings.Split(arg, ",") {
			if strings.TrimSpace(test) != "" {
				cliOptions.Tests = append(cliOptions.Tests, strings.TrimSpace(test))
			}
		}
	}

	// verification of test cases passed through cli.
	for _, testId := range cliOptions.Tests {
		if !testengine.IsTestCase(strings.ToUpper(testId)) {
			return fmt.Errorf("The %[2]s is not a valid test case id. Usee '%[1]s list' to list valid test cases. Aborting!", os.Args[0], testId)
		}
	}

	return nil
}
