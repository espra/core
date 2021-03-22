// Public Domain (-) 2018-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

package process

import (
	"context"
	"os/exec"
	"runtime"
	"syscall"
	"testing"
)

func TestReapOrphans(t *testing.T) {
	if runtime.GOOS != "linux" {
		ReapOrphans()
		return
	}
	testMode = true
	cmd := exec.Command("sleep", "100")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Unexpected error when trying to run `sleep 100`: %s", err)
	}
	if more := ReapOrphans(); !more {
		t.Fatalf("Failed to find unterminated child process when calling ReapOrphans")
	}
	syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
	if more := ReapOrphans(); more {
		t.Fatalf("Unterminated child process encountered when calling ReapOrphans")
	}
}

func TestRunReaper(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go RunReaper(ctx)
	defer cancel()
	if runtime.GOOS != "linux" {
		return
	}
	testMode = true
	cmd := exec.Command("sleep", "100")
	if err := cmd.Start(); err != nil {
		t.Fatalf("Unexpected error when trying to run `sleep 100`: %s", err)
	}
	syscall.Kill(cmd.Process.Pid, syscall.SIGTERM)
	<-testSig
}
