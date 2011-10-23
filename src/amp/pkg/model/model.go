// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package model

type Item struct { /* Parent: User */
	Created   int64
	Head      bool
	Index     []string /* From, By, []To, []About, Aspect, Refs, Unpacked Value, Words in Message */
	Message   string
	Parent1   string
	Parent2   string
	Parent3   string
	Value     []byte
	ValueType int
	Version   int
}

type Pointer struct { /* Parent: User */
	Name string
	Ref  string
	Type int
}

type User struct {
	Avatar     []byte // storage key
	AvatarType int    /* Gravatar, Facebook, etc. */
	Email      string
	Externals  [][]byte
	Created    int64
	Modified   int64
	Passphrase []byte
	Phone      []byte
	PrefNick   []byte
	Status     int /* Banned */
	Validated  int
	Version    int
}

type ExternalAccount struct { /* Parent: User, Key: Domain-UserID */
	Created     int64
	Domain      string /* twitter, facebook, etc. */
	DisplayName string
	Modified    int64
	Scope       int /* read, read/write, etc. */
	Token       []byte
	Type        int /* OpenID, OAuth 1.0, OAuth 2.0, etc. */
	UserID      string
	Username    string
	Version     int
}

type ExternalProfile struct { /* Parent: ExternalAccount */
	Data []byte
}

type ExternalConnections struct { /* Parent: ExternalAccount */
	Peers [][]byte /* ??? */
}

type UserStatus struct {
	Payment int
}

type OAuth struct {

}

type Handshake struct {

}

type Hook struct {

}

type Token struct {

}

type Space struct {

}

type Subscription struct {

}

type Log struct {
	Created   int64
	User      string
	Message   []byte
	Level     int
	Error     bool
	ErrorType string
	Traceback string
}
