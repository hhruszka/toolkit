package testengine

import (
	"strings"
	"time"
)

// ExitCode is an enumeration of possible exit codes with descriptive names.
// It provides a more idiomatic way to refer to exit codes within the Go application.
type RetCode int

const (
	ExecutionCanceled RetCode = iota - 3
	ExecutionTimeOut
	InternalAppError
	Success
	GeneralError
	IncorrectUsage
	CommandCannotExecute  = 126
	CommandNotFound       = 127
	InvalidArgumentToExit = 128
	// Skips to specific values after the iota increment
	ScriptTerminatedByControlC RetCode = 130
	ExitStatusOutOfRange       RetCode = 255
	// Signal based exit codes (128+n)
	FatalErrorSignal1 RetCode = 129
	// FatalErrorSignal2 is omitted as it overlaps with ScriptTerminatedByControlC
	FatalErrorSignal3  RetCode = 131
	FatalErrorSignal4  RetCode = 132
	FatalErrorSignal5  RetCode = 133
	FatalErrorSignal6  RetCode = 134
	FatalErrorSignal7  RetCode = 135
	FatalErrorSignal8  RetCode = 136
	FatalErrorSignal9  RetCode = 137
	FatalErrorSignal10 RetCode = 138
	FatalErrorSignal11 RetCode = 139
	FatalErrorSignal12 RetCode = 140
	FatalErrorSignal13 RetCode = 141
	FatalErrorSignal14 RetCode = 142
	FatalErrorSignal15 RetCode = 143
)

type ExecutionStatus struct {
	RetCode  RetCode   `json:"RetCode"`
	Error    []string  `json:"Error"`
	Stdout   []string  `json:"Stdout"`
	Stderr   []string  `json:"Stderr"`
	ExecTime time.Time `json:"ExecTime"`
	Data     any       `json:"Data"`
}

func NewExecutionStatus(exitStatus RetCode, error string, stdout string, stderr string, execTime time.Time) *ExecutionStatus {
	return &ExecutionStatus{RetCode: exitStatus, Error: strings.Split(error, "\n"), Stdout: strings.Split(stdout, "\n"), Stderr: strings.Split(stderr, "\n"), ExecTime: execTime}
}

func NewExecutionStatusWithData(exitStatus RetCode, error string, stdout string, stderr string, execTime time.Time, data any) *ExecutionStatus {
	st := NewExecutionStatus(exitStatus, error, stdout, stderr, execTime)
	st.SetData(data)
	return st
}

func (es *ExecutionStatus) SetData(data any) {
	es.Data = data
}

func (es *ExecutionStatus) GetData() any {
	return es.Data
}

type TestResultStatus int

const (
	Passed TestResultStatus = iota
	Failed
	Timeout
	NotApplicable
)

type TestResults struct {
	TestId     string
	ExecStatus *ExecutionStatus
	Status     TestResultStatus
	Error      string
	Data       any
}

type RegexMatches struct {
	pattern *Pattern
	matches []string
}

type Probability int

const (
	VeryUnlikely Probability = iota
	Unlikely
	Possible
	Likely
	VeryLikely
)

type Secret struct {
	SecretType  string
	SecretValue string
	Likelihood  Probability
	LineNumber  int
}

type SecretScanResults struct {
	file    string
	secrets map[int]Secret
}
