package main

import (
	"flag"
	"fmt"
	"os"

	"claude-notifications-win/cmd"
)

const version = "1.0.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmdName := os.Args[1]

	switch cmdName {
	case "version":
		fmt.Println(version)
	case "stop":
		flag.Parse()
		if err := cmd.HandleStopHook(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
	case "permission":
		flag.Parse()
		if err := cmd.HandlePermissionHook(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: notify.exe <command>")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  version    Show version")
	fmt.Fprintln(os.Stderr, "  stop       Handle task completion notification")
	fmt.Fprintln(os.Stderr, "  permission Handle permission prompt notification")
}
