// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package model

type Item struct { /* Parent: User */
	Created   int64
	Head      bool
	Index     [][]byte /* From, By, []To, []About, Aspect, Refs, Unpacked Value, Words in Message */
	Message   []byte
	Parent1   string
	Parent2   string
	Parent3   string
	Value     []byte
	ValueType int32
	Version   int32
}

type Pointer struct { /* Parent: User */
	Name string
	Ref  string
	Type int32
}

type User struct {
	Avatar     []byte // storage key
	AvatarType int32  /* Gravatar, Facebook, etc. */
	Email      string
	Externals  [][]byte
	Created    int64
	Modified   int64
	Passphrase []byte
	Phone      []byte
	PrefNick   []byte
	Status     int32 /* Banned */
	Validated  int32
	Version    int32
}

type ExternalAccount struct { /* Parent: User, Key: Domain-UserID */
	Created     int64
	Domain      string /* twitter, facebook, etc. */
	DisplayName string
	Modified    int64
	Scope       int32 /* read, read/write, etc. */
	Token       []byte
	Type        int32 /* OpenID, OAuth 1.0, OAuth 2.0, etc. */
	UserID      string
	Username    string
	Version     int32
}

type ExternalProfile struct { /* Parent: ExternalAccount */
	Data []byte
}

type ExternalConnections struct {  /* Parent: ExternalAccount */
	Peers [][]byte /* ??? */
}

type UserStatus struct {
	Payment int32
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
