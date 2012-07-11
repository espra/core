// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package master

type Client struct {
	nodes      []*Node
	headMaster int
}

func (client *Client) Init() error {
	// Who is the head master?
}

func (client *Client) Get(key string) {

}

func (client *Client) DirtyGet(key string) {

}

func (client *Client) GetLease(from, to string) {

}

func (client *Client) RenewLease(key string) {

}

type Node struct {
	Address   string
	PublicKey []byte
}

func NewClient(nodes []*Node) (client *Client, err error) {
	client := &Client{nodes}
	err = client.Init()
	return
}
