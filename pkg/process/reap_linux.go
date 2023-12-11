// Public Domain (-) 2018-present, The Espra Core Authors.
// See the Espra Core UNLICENSE file for details.

package process

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sys/unix"
)

func reap() bool {
	status := syscall.WaitStatus(0)
	for {
		pid, err := syscall.Wait4(-1, &status, unix.WNOHANG, nil)
		if pid == 0 {
			return true
		}
		if pid == -1 && err == syscall.ECHILD {
			return false
		}
	}
}

func runReaper(ctx context.Context) {
	if os.Getpid() != 1 {
		unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, uintptr(1), 0, 0, 0)
	}
	notifier := make(chan os.Signal, 4096)
	signal.Notify(notifier, syscall.SIGCHLD)
outer:
	for {
		select {
		case <-notifier:
			reap()
			if testMode {
				testSig <- struct{}{}
			}
		case <-ctx.Done():
			signal.Stop(notifier)
			break outer
		}
	}
}
