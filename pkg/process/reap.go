// Public Domain (-) 2018-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// +build !linux

package process

import (
	"context"
)

func reap() bool {
	return false
}

func runReaper(ctx context.Context) {
}
