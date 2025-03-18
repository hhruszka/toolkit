package log

import (
	"fmt"
	"os"
	"sync"
)

// global variables
var (
	logBuffer      chan string = make(chan string, 100)
	logWg          sync.WaitGroup
	loggingEnabled bool
)

func init() {
	loggingEnabled = false
	Start()
}

func Start() {
	logWg.Add(1)
	go func() {
		defer logWg.Done()
		for msg := range logBuffer {
			fmt.Fprint(os.Stderr, msg)
		}
	}()
}

func LoggingEnabled() {
	loggingEnabled = true
}

func LoggingDisabled() {
	loggingEnabled = false
}

func Logf(format string, a ...interface{}) string {
	logMsg := fmt.Sprintf(format, a...)
	if loggingEnabled {
		logBuffer <- logMsg
	}
	return logMsg
}

func Logln(a ...interface{}) string {
	logMsg := fmt.Sprintln(a...)
	if loggingEnabled {
		logBuffer <- fmt.Sprintln(a...)
	}
	return logMsg
}

func Fatal(a ...interface{}) {
	// This defer-recover pattern catches the panic that occurs when closing an already closed channel.
	defer func() {
		// recover stops the panic and returns the value that was passed to the call of panic.
		if recover() != nil {
			fmt.Println("Attempted to close an already closed channel")
		}
	}()

	logMsg := fmt.Sprintln(a...)
	if loggingEnabled {
		logBuffer <- fmt.Sprintln(a...)
		close(logBuffer)
		logWg.Wait()
	}
	panic(logMsg)
}

func Exit() {
	// This defer-recover pattern catches the panic that occurs when closing an already closed channel.
	defer func() {
		// recover stops the panic and returns the value that was passed to the call of panic.
		if recover() != nil {
			fmt.Println("Attempted to close an already closed channel")
		}
	}()

	close(logBuffer)
	logWg.Wait()
}
