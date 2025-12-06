package execx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fengjx/go-halo/osx"
)

// from https://github.com/zeromicro/go-zero/blob/master/tools/goctl/rpc/execx/execx.go

// Run provides the execution of shell scripts in golang,
// which can support macOS, Windows, and Linux operating systems.
// Other operating systems are currently not supported
func Run(ctx context.Context, arg, dir string, in ...io.Reader) (string, error) {
	goos := runtime.GOOS
	var cmd *exec.Cmd
	switch goos {
	case osx.OsMac, osx.OsLinux:
		cmd = exec.CommandContext(ctx, "sh", "-c", arg)
	case osx.OsWindows:
		cmd = exec.CommandContext(ctx, "cmd.exe", "/c", arg)
	default:
		return "", fmt.Errorf("unexpected os: %v", goos)
	}
	if len(dir) > 0 {
		cmd.Dir = dir
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	if len(in) > 0 {
		cmd.Stdin = in[0]
	}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	err := cmd.Run()
	if err != nil {
		if stderr.Len() > 0 {
			return "", errors.New(strings.TrimSuffix(stderr.String(), osx.NL))
		}
		return "", err
	}
	return strings.TrimSuffix(stdout.String(), osx.NL), nil
}

// RunWithConsole provides the execution of shell scripts in golang,
// which can support macOS, Windows, and Linux operating systems.
// Other operating systems are currently not supported
func RunWithConsole(cmdLine, dir string) error {
	return RunWithConsoleContext(context.Background(), cmdLine, dir)
}

// RunWithConsoleContext provides the execution of shell scripts with context support,
// which can support macOS, Windows, and Linux operating systems.
// Other operating systems are currently not supported
func RunWithConsoleContext(ctx context.Context, cmdLine, dir string) error {
	goos := runtime.GOOS
	var cmd *exec.Cmd
	switch goos {
	case osx.OsMac, osx.OsLinux:
		cmd = exec.CommandContext(ctx, "sh", "-c", cmdLine)
	case osx.OsWindows:
		cmd = exec.CommandContext(ctx, "cmd.exe", "/c", cmdLine)
	default:
		return fmt.Errorf("unexpected os: %v", goos)
	}
	if len(dir) > 0 {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if err != nil {
		// 如果是 context 取消导致的错误，返回 context.Canceled
		if errors.Is(ctx.Err(), context.Canceled) {
			return context.Canceled
		}
		return err
	}
	return nil
}

// WrapCmd wraps the command and arguments into a single string
func WrapCmd(cmd string, args []string) string {
	return strings.Join(append([]string{cmd}, args...), " ")
}
