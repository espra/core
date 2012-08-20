// Public Domain (-) 2012 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package httputil

import (
	"amp/structure"
	"net/http"
	"net/textproto"
	"sort"
	"strconv"
	"strings"
)

type AcceptOption struct {
	identity     bool
	order        int
	metaPrefix   string
	metaWildcard bool
	value        string
	weight       float64
	wildcard     bool
}

type AcceptOptions []*AcceptOption

func (s AcceptOptions) Len() int      { return len(s) }
func (s AcceptOptions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s AcceptOptions) Less(i, j int) bool {
	if s[i].weight > s[j].weight {
		return true
	}
	if s[i].wildcard || s[j].wildcard || s[i].metaWildcard || s[j].metaWildcard {
		if s[i].wildcard {
			return false
		} else {
			if s[j].wildcard {
				return true
			}
			if s[i].metaWildcard {
				if s[j].metaWildcard {
					return s[i].metaPrefix < s[j].metaPrefix
				}
				return false
			}
			if s[j].metaWildcard {
				return true
			}
		}
	}
	return s[i].order < s[j].order
}

type Acceptable struct {
	codings bool
	opts    AcceptOptions
}

func (a *Acceptable) Accepts(value string) bool {
	if len(a.opts) == 0 {
		if a.codings {
			if value == "identity" {
				return true
			}
			return false
		} else {
			return true
		}
	}
	for _, opt := range a.opts {
		if opt.weight == 0 {
			return false
		}
		if opt.identity {
			if value == "identity" {
				return true
			}
			return false
		}
		if opt.wildcard {
			return true
		}
		if opt.value == value {
			return true
		}
		if opt.metaWildcard && strings.HasPrefix(value, opt.metaPrefix) {
			return true
		}
	}
	return false
}

func (a *Acceptable) FindPreferred(values ...string) []string {
	matches := []string{}
	if len(a.opts) == 0 {
		if a.codings {
			for _, value := range values {
				if value == "identity" {
					return []string{"identity"}
				}
			}
			return matches
		}
		for _, value := range values {
			matches = append(matches, value)
		}
		return matches
	}
	for _, opt := range a.opts {
		if opt.weight == 0 {
			break
		}
		if opt.wildcard {
			for _, value := range values {
				if !structure.InStringSlice(matches, value) {
					matches = append(matches, value)
				}
			}
			return matches
		}
		for _, value := range values {
			if opt.value == value {
				if !structure.InStringSlice(matches, value) {
					matches = append(matches, value)
				}
			} else if opt.metaWildcard && strings.HasPrefix(value, opt.metaPrefix) && !structure.InStringSlice(matches, value) {
				matches = append(matches, value)
			}
		}
	}
	return matches
}

// Parse handles special HTTP header fields like Accept-Encoding and returns a
// queryable object.
func Parse(r *http.Request, key string) *Acceptable {
	key = textproto.CanonicalMIMEHeaderKey(key)
	value := r.Header.Get(key)
	a := &Acceptable{}
	if value == "" {
		return a
	}
	var err error
	for idx, part := range strings.Split(value, ",") {
		parts := strings.Split(part, ";")
		weight := 1.0
		if len(parts) >= 2 {
			for _, qvalue := range parts[1:] {
				if len(qvalue) >= 3 && qvalue[:2] == "q=" {
					weight, err = strconv.ParseFloat(qvalue[2:], 64)
					if err != nil {
						continue
					}
					break
				}
			}
		}
		part = strings.TrimSpace(parts[0])
		opt := &AcceptOption{
			order:  idx,
			value:  part,
			weight: weight,
		}
		switch key {
		case "Accept":
			if part == "*/*" {
				opt.wildcard = true
			} else if strings.HasSuffix(part, "/*") {
				opt.metaPrefix = part[:len(part)-1]
				opt.metaWildcard = strings.Count(opt.metaPrefix, "/") == 1
			}
		case "Accept-Charset":
			if part == "*" {
				opt.wildcard = true
			}
		case "Accept-Encoding":
			if part == "*" {
				opt.wildcard = true
			} else if part == "identity" {
				opt.identity = true
			}
		case "Accept-Language":
			if part == "*" {
				opt.wildcard = true
			} else if !strings.Contains(part, "-") {
				opt.metaPrefix = part + "-"
				opt.metaWildcard = true
			}
		}
		a.opts = append(a.opts, opt)
	}
	if key == "Accept-Charset" {
		iso88591 := true
		for _, opt := range a.opts {
			if opt.value == "iso-8859-1" {
				iso88591 = false
				break
			}
		}
		if iso88591 {
			a.opts = append(a.opts, &AcceptOption{
				order:  len(a.opts),
				value:  "iso-8859-1",
				weight: 1.0,
			})
		}
	} else if key == "Accept-Encoding" {
		a.codings = true
	}
	if len(a.opts) > 0 {
		sort.Sort(a.opts)
	}
	return a
}
