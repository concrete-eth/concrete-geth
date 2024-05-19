package main

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/fatih/color"
)

var (
	green  = color.New(color.FgGreen)
	red    = color.New(color.FgRed)
	gray   = color.New(color.FgHiBlack)
	yellow = color.New(color.FgYellow)
	bold   = color.New(color.Bold)
)

func logInfo(format string, a ...any) {
	fmt.Printf(format+"\n", a...)
}

func logDebug(format string, a ...any) {
	gray.Printf(format+"\n", a...)
}

func logWarning(warning string) {
	yellow.Fprint(os.Stderr, "Warning: ")
	fmt.Fprintln(os.Stderr, warning)
}

func logError(err error, context bool) {
	fmt.Fprintln(os.Stderr, "Error:")
	red.Fprintln(os.Stderr, err)
	if context {
		fmt.Fprintln(os.Stderr, "Context:")
		gray.Fprintln(os.Stderr, string(debug.Stack()))
	}
}

func logFatal(err error) {
	logError(err, true)
	os.Exit(1)
}

func logFatalNoContext(err error) {
	logError(err, false)
	os.Exit(1)
}
