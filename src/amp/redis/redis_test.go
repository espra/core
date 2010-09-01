// No Copyright (-) 2010 The Ampify Authors. This file is under the
// Public Domain license that can be found in the root LICENSE file.

package redis

import (
	"fmt"
	"testing"
)

func TestVersion(t *testing.T) {

	client := Client("localhost:6379")
	fmt.Print(client)

	client = Client()
	fmt.Print(client)

	client = Client("unix:/foo/bar")
	fmt.Print(client)

}
