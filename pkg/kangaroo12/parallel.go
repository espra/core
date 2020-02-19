// Public Domain (-) 2020-present, The Core Authors.
// See the Core UNLICENSE file for details.

package kangaroo12

import (
	"io"
	"sync"
)

// Parallel processes the data with the given number of threads and returns a
// Reader for use as an XOF. Use a nil customIdent and read the first 32 bytes
// from the Reader to use as a traditional hash function.
func Parallel(threads int, customIdent []byte, data []byte) io.Reader {
	data = append(data, customIdent...)
	data = append(data, encodeLength(len(customIdent))...)
	count := (len(data) + chunkSize - 1) / chunkSize
	meta := &sponge{}
	meta.init()
	if count == 1 {
		meta.absorb(data)
		meta.finalize(0x07)
		return meta
	}
	meta.absorb(data[:chunkSize])
	data = data[chunkSize:]
	limit := count - 1
	last := limit - 1
	chunks := make([]byte, 32*limit)
	var wg sync.WaitGroup
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func(worker int) {
			chunk := &sponge{}
			chunk.init()
			for i := worker; i < limit; i += threads {
				start := i * chunkSize
				if i == last {
					chunk.absorb(data[start:])
				} else {
					chunk.absorb(data[start : start+chunkSize])
				}
				chunk.finalize(0x0b)
				start = i * 32
				chunk.squeeze(chunks[start : start+32])
				chunk.reset()
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
	meta.absorb([]byte{3, 0, 0, 0, 0, 0, 0, 0})
	meta.absorb(chunks)
	meta.absorb(encodeLength(limit))
	meta.absorb([]byte{0xff, 0xff})
	meta.finalize(0x06)
	return meta
}
