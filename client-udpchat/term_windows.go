//go:build windows
// +build windows

package main

import (
    "golang.org/x/sys/windows"
    "os"
)

func enableVirtualTerminalProcessing() {
    stdout := windows.Handle(os.Stdout.Fd())
    var mode uint32
    if err := windows.GetConsoleMode(stdout, &mode); err == nil {
        mode |= windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
        _ = windows.SetConsoleMode(stdout, mode)
    }
}

