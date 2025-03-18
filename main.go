package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"linuxtester/log"
	"linuxtester/testengine"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	// app name
	appName string = filepath.Base(os.Args[0])
	// AppVersion - app version
	AppVersion string
)

// CLI options variables
var (
	timeout    int64
	debug      bool
	format     string
	list       bool
	info       bool
	version    bool
	failedonly bool
	detailed   bool
	tests      []string
)

// App global variables
var (
	HostName string
	UserName string
	UserID   int
)

func printTestList() error {
	switch format {
	case "text":
		tests := testengine.PrintTestCasesOnly()
		fmt.Println(tests.String())
	case "json":
		jsonBuff, err := json.MarshalIndent(testengine.TestCases, "", "    ")
		if err != nil {
			return fmt.Errorf("Internal application error: %s\n", err.Error())
		}
		fmt.Println(string((jsonBuff)))
	}
	return nil
}

func result(res bool) string {
	return map[bool]string{true: color.GreenString("PASSED"), false: color.RedString("FAILED")}[res]
}

func filterFailedOnly(userTestsResults []*testengine.AccountTestStatus) {
	if failedonly {
		for _, testsResults := range userTestsResults {
			for testId, test := range testsResults.ExecTestStatuses {
				if test.Status {
					delete(testsResults.ExecTestStatuses, testId)
				}
			}
		}
	}
}

func generateReport(userTestsResults []*testengine.AccountTestStatus) error {
	var buf bytes.Buffer

	switch format {
	case "json":
		jsonBuff, err := json.MarshalIndent(userTestsResults, "", "    ")
		if err != nil {
			return fmt.Errorf("Internal application error: %s\n", err.Error())
		}
		fmt.Println(string((jsonBuff)))
	case "text":
		for _, testsResults := range userTestsResults {
			var sortedTestIds []string

			for testId, _ := range testsResults.ExecTestStatuses {
				sortedTestIds = append(sortedTestIds, testId)
			}
			sort.Strings(sortedTestIds)

			t := table.NewWriter()
			t.SetOutputMirror(&buf)
			t.SetTitle("Host: %s\nUser: %s\nDate: %s", HostName, testsResults.UserName, time.Now().Format("Jan 02, 2006 3:04 PM"))
			t.SetAllowedRowLength(140)

			for _, testId := range sortedTestIds {
				//if failedonly && testsResults.ExecTestStatuses[testId].Status {
				//	//skip passed test cases -> PASSED when container.Results[testId].Status == true
				//	continue
				//}

				if testengine.AllTestCases[testId].IsTest {
					t.AppendRows([]table.Row{{testId, testengine.AllTestCases[testId].Abstract, result(testengine.AllTestCases[testId].ResultFunc(testsResults.ExecTestStatuses[testId].ExecStatus))}})
					t.AppendSeparator()
				}
			}
			t.Render()
		}
		fmt.Println(buf.String())
	}
	return nil
}

func wrap(text []string, n int) string {
	var buffer bytes.Buffer

	for _, str := range text {
		wrapLen := n - 1
		strLen := len(str) - 1
		for idx, c := range str {
			buffer.WriteRune(c)
			if idx%n == wrapLen && idx != strLen {
				buffer.WriteRune('\n')
			}
		}
		buffer.WriteRune('\n')
	}
	return buffer.String()
}

func generateReportDetailed(userTestsResults []*testengine.AccountTestStatus) error {
	var report bytes.Buffer
	for _, testsResults := range userTestsResults {
		var buf bytes.Buffer
		var sortedTestIds []string
		for testId, _ := range testsResults.ExecTestStatuses {
			sortedTestIds = append(sortedTestIds, testId)
		}
		sort.Strings(sortedTestIds)

		t := table.NewWriter()
		t.SetOutputMirror(&buf)
		t.SetTitle("Host: %s\nUser: %s\nDate: %s", HostName, testsResults.UserName, time.Now().Format("Jan 02, 2006 3:04 PM"))
		t.SetAllowedRowLength(140)

		for _, testId := range sortedTestIds {
			if testengine.AllTestCases[testId].IsTest {
				t.AppendRow(nil)
				t.AppendSeparator()
				t.AppendRows([]table.Row{{testId, testengine.AllTestCases[testId].Abstract, result(testengine.AllTestCases[testId].ResultFunc(testsResults.ExecTestStatuses[testId].ExecStatus))}})
				t.AppendSeparator()
				t.SetOutputMirror(&buf)
				t.AppendRow(table.Row{"Exit status:", testsResults.ExecTestStatuses[testId].ExecStatus.RetCode})
				t.AppendSeparator()
				t.AppendRow(table.Row{"Error", wrap(testsResults.ExecTestStatuses[testId].ExecStatus.Error, 100)})
				t.AppendSeparator()
				t.AppendRow(table.Row{"Stdout:", wrap(testsResults.ExecTestStatuses[testId].ExecStatus.Stdout, 100)})
				t.AppendSeparator()
				t.AppendRow(table.Row{"Stderr:", wrap(testsResults.ExecTestStatuses[testId].ExecStatus.Stderr, 100)})
				t.AppendSeparator()
			}
		}
		t.Render()
		fmt.Fprintf(&report, buf.String())
	}

	return SaveReportToFile(report.Bytes())
}

func SaveReportToFile(report []byte) error {
	filePath := filepath.Join("./report-bptvnftester-" + time.Now().Format("2006-01-02_15-04-05_MST") + ".txt")
	if err := os.WriteFile(filePath, report, 0444); err != nil {
		return fmt.Errorf("failed to save to file due to: %w", err)
	}
	_, _ = fmt.Fprintln(os.Stderr, "Report saved successfully:", filePath)
	return nil
}

func run(args []string) error {
	defer log.Exit()

	if version {
		fmt.Println(appName, AppVersion)
		return nil
	}

	if list {
		printTestList()
		return nil
	}

	if info {
		//return printInfo(k8sExecClient)
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
			return fmt.Errorf("The %[2]s is not a valid test case id. Usee '%[1]s -l' or '%[1]s --list' to list valid test cases. Aborting!", os.Args[0], testId)
		}
	}

	// debug is a cli flag turning logging on. Logs are output to standard error
	// it cannot be part of the below switch because it just a flag.
	if debug {
		log.LoggingEnabled()
	}

	//getUsers(k8sExecClient, containers)
	var testsResults []*testengine.AccountTestStatus

	users := testengine.GetUsers()
	if os.Getuid() == 0 && len(users) > 0 {
		testsResults = testengine.RunAccountTests(tests, users, time.Duration(timeout)*time.Second)
	} else {
		testsResults = append(testsResults, testengine.NewAccountTestStatus(UserName, testengine.RunTests(tests, "sh", nil, time.Duration(timeout)*time.Second)))
	}

	if failedonly {
		filterFailedOnly(testsResults)
	}

	if detailed && format != "json" {
		return generateReportDetailed(testsResults)
	}
	return generateReport(testsResults)
}

var cmd = &cobra.Command{
	Use:           appName + " [flags] [test ids]",
	Short:         appName + " is a command line application that executes baseline penetration test cases on hosts (VMs, k8s nodes etc.)",
	Long:          ``,
	SilenceErrors: true,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if timeout < 0 {
			_ = cmd.Usage()
			return fmt.Errorf("\nWrong value for timeout option!\ntimeout cannot be a negative value, aborting")
		}

		if format != "text" && format != "json" {
			_ = cmd.Usage()
			return fmt.Errorf("\nWrong value for output format option!\nthe %s is not a valid report format for the output option (-o or --output), aborting", format)
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		return run(args)
	},
}

func init() {
	var err error

	UserName = strconv.Itoa(os.Getuid())
	userInfo, err := user.Current()
	if err == nil {
		UserName = userInfo.Name
	}

	HostName, _ = os.Hostname()

	cmd.Flags().BoolVarP(&list, "list", "l", false, "list all test cases")
	cmd.Flags().BoolVarP(&version, "version", "v", false, "prints "+appName+" version")
	cmd.Flags().BoolVarP(&failedonly, "failed-only", "", false, "create a report with failed only test cases")
	cmd.Flags().BoolVarP(&detailed, "detailed-report", "", false, "create a report with execution details for ticketing")
	cmd.Flags().StringVarP(&format, "output", "o", "text", "report format: text, or json")
	cmd.Flags().Int64VarP(&timeout, "timeout", "t", 300, "timeout in seconds")
	cmd.Flags().BoolVar(&debug, "debug", false, "debug flag")

	// Disable automatic printing of usage when an error occurs
	cmd.SilenceUsage = true

	// support for '--'
	cmd.Flags().SetInterspersed(false)

	// Custom PreRunE to check for parse errors
	cmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// This checks for any parse errors
		if err := cmd.ParseFlags(args); err != nil {
			return err
		}
		return nil
	}

	cmd.SetFlagErrorFunc(func(c *cobra.Command, err error) error {
		// When a non-existing option is invoked, print the usage
		if err := c.Usage(); err != nil {
			fmt.Fprintf(os.Stderr, "Error printing usage: %v\n", err)
		}
		// Return the original error to stop execution
		return err
	})

}

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}
}
