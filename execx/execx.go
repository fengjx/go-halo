package execx

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fengjx/go-halo/osx"
)

// from https://github.com/zeromicro/go-zero/blob/master/tools/goctl/rpc/execx/execx.go

// RunFunc defines a function type of Run function
type RunFunc func(string, string, ...*bytes.Buffer) (string, error)

// Run provides the execution of shell scripts in golang,
// which can support macOS, Windows, and Linux operating systems.
// Other operating systems are currently not supported
func Run(arg, dir string, in ...*bytes.Buffer) (string, error) {
	goos := runtime.GOOS
	var cmd *exec.Cmd
	switch goos {
	case osx.OsMac, osx.OsLinux:
		cmd = exec.Command("sh", "-c", arg)
	case osx.OsWindows:
		cmd = exec.Command("cmd.exe", "/c", arg)
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

// WrapCmd wraps the command and arguments into a single string
func WrapCmd(cmd string, args []string) string {
	return strings.Join(append([]string{cmd}, args...), " ")
}
