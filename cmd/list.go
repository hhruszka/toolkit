package cmd

import (
	"bptvnftester/testengine"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
)

func NewCmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all test cases to standard output",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return printTestList("text")
		},
	}

	return cmd
}

func printTestList(format string) error {
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
