// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package blake3

import (
	"dappui.com/cmd/genasm/pkg"
)

func init() {
	pkg.Register("blake3", &pkg.Entry{
		File:      "hash_avx2",
		Generator: genAVX2,
		Stub:      true,
	})
}
