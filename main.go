package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"syscall"
	"unsafe"

	"claude-notifications-win/src/cmd"
)

var version = "dev" // Set via -ldflags at build time

func main() {
	if len(os.Args) < 2 {
		if isInteractive() {
			// 双击运行：直接进入配置向导（引导填飞书 webhook 等），
			// 结束后等待回车，避免控制台窗口闪关看不到结果
			exitCode := 0
			if err := cmd.RunInit(); err != nil {
				fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
				exitCode = 1
			}
			fmt.Println()
			fmt.Print("按回车键退出...")
			bufio.NewReader(os.Stdin).ReadString('\n')
			os.Exit(exitCode)
		}
		// 非交互（管道/脚本调用）：打印用法后退出
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
	case "init":
		if err := cmd.RunInit(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
	case "test":
		if err := cmd.RunTest(); err != nil {
			fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmdName)
		printUsage()
		os.Exit(1)
	}
}

var procGetConsoleMode = syscall.NewLazyDLL("kernel32.dll").NewProc("GetConsoleMode")

// isInteractive 判断 stdin 是否为真实控制台（双击运行场景）。
// 不能用 ModeCharDevice：Windows 上 NUL（/dev/null）也是字符设备；
// GetConsoleMode 只对真实控制台句柄成功，管道 / NUL / 重定向均失败。
func isInteractive() bool {
	var mode uint32
	r, _, _ := procGetConsoleMode.Call(os.Stdin.Fd(), uintptr(unsafe.Pointer(&mode)))
	return r != 0
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage: .\\notify.exe <command>")
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  version    Show version")
	fmt.Fprintln(os.Stderr, "  stop       Handle task completion notification")
	fmt.Fprintln(os.Stderr, "  permission Handle permission prompt notification")
	fmt.Fprintln(os.Stderr, "  init       Interactive setup wizard")
	fmt.Fprintln(os.Stderr, "  test       Test all notification channels")
}
