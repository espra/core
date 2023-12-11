// Public Domain (-) 2018-present, The Espra Core Authors.
// See the Espra Core UNLICENSE file for details.

//go:build !linux

package process

import (
	"context"
)

func reap() bool {
	return false
}

func runReaper(ctx context.Context) {
}
