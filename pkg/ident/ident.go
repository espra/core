// Public Domain (-) 2018-present, The Web4 Authors.
// See the Web4 UNLICENSE file for details.

// Package ident provides support for converting identifiers between different
// naming conventions.
package ident

import (
	"bytes"
	"fmt"
)

// Parts represents the normalized elements of an identifier.
type Parts [][]byte

func (p Parts) String() string {
	return string(bytes.Join(p, []byte{','}))
}

// ToCamel converts the identifier into a camelCased string.
func (p Parts) ToCamel() string {
	var out []byte
	for idx, part := range p {
		if idx == 0 {
			out = append(out, bytes.ToLower(part)...)
		} else {
			out = append(out, part...)
		}
	}
	return string(out)
}

// ToKebab converts the identifier into a kebab-cased string.
func (p Parts) ToKebab() string {
	var out []byte
	for idx, part := range p {
		if idx != 0 {
			out = append(out, '-')
		}
		out = append(out, bytes.ToLower(part)...)
	}
	return string(out)
}

// ToPascal converts the identifier into a PascalCased string.
func (p Parts) ToPascal() string {
	var out []byte
	for _, part := range p {
		out = append(out, part...)
	}
	return string(out)
}

// ToScreamingSnake converts the identifier into a SCREAMING_SNAKE_CASED string.
func (p Parts) ToScreamingSnake() string {
	var out []byte
	for idx, part := range p {
		if idx != 0 {
			out = append(out, '_')
		}
		out = append(out, bytes.ToUpper(part)...)
	}
	return string(out)
}

// ToSnake converts the identifier into a snake_cased string.
func (p Parts) ToSnake() string {
	var out []byte
	for idx, part := range p {
		if idx != 0 {
			out = append(out, '_')
		}
		out = append(out, bytes.ToLower(part)...)
	}
	return string(out)
}

// add appends parts from the given element. It looks for runs of initialisms
// like "HTTPAPIs" and adds them as separate parts, i.e. "HTTP" and "APIs". Once
// all initialisms are detected, the remaining element is added as a single
// part.
func (p Parts) add(elem []byte) Parts {
	// Try to match an initialism exactly.
	if special, ok := mapping[string(bytes.ToUpper(elem))]; ok {
		return append(p, []byte(special))
	}
	// Try to find the longest initialism matches from the start.
	for len(elem) > 0 {
		match := ""
		pos := -1
		for i := 0; i <= len(elem); i++ {
			if special, ok := mapping[string(bytes.ToUpper(elem[:i]))]; ok {
				match = special
				pos = i
			}
		}
		if pos == -1 {
			p = append(p, elem)
			break
		}
		p = append(p, []byte(match))
		elem = elem[pos:]
	}
	return p
}

// tryAdd attempts to add parts from the given element. If any initialisms are
// found, they are added in canonical form.
func (p Parts) tryAdd(elem []byte) (Parts, []byte) {
	var nelem []byte
	// Try to match an initialism exactly.
	if special, ok := mapping[string(bytes.ToUpper(elem))]; ok {
		return append(p, []byte(special)), nil
	}
	// Try to match an initialism from the end for the longest identifier with a
	// non-uppercase suffix.
	last := ""
	pos := -1
	for i := len(elem) - 1; i >= 0; i-- {
		if special, ok := mapping[string(bytes.ToUpper(elem[i:]))]; ok {
			last = special
			pos = i
		}
	}
	if pos == -1 {
		// NOTE(tav): The given elem must be at least 2 characters long. The
		// code in FromPascal currently ensures this to be the case.
		nelem = elem[len(elem)-2:]
		elem = elem[:len(elem)-2]
	} else {
		elem = elem[:pos]
	}
	p = p.add(elem)
	if len(last) > 0 {
		p = append(p, []byte(last))
	}
	return p, nelem
}

// FromCamel parses the given camelCased identifier into its parts.
func FromCamel(ident string) Parts {
	var parts Parts
	i := 0
	for ; i < len(ident); i++ {
		char := ident[i]
		if char >= 'A' && char <= 'Z' {
			break
		}
	}
	parts = append(parts, normalize([]byte(ident[:i])))
	// NOTE(tav): The error must be nil, as ident must be empty or start on an
	// uppercase character, per the break clause above.
	elems, _ := FromPascal(ident[i:])
	return append(parts, elems...)
}

// FromKebab parses the given kebab-cased identifier into its parts.
func FromKebab(ident string) Parts {
	var (
		elem  []byte
		parts Parts
	)
	for i := 0; i < len(ident); i++ {
		char := ident[i]
		if char == '-' {
			if len(elem) == 0 {
				continue
			}
			parts = append(parts, normalize(bytes.ToLower(elem)))
			elem = []byte{}
		} else {
			elem = append(elem, char)
		}
	}
	if len(elem) > 0 {
		parts = append(parts, normalize(bytes.ToLower(elem)))
	}
	return parts
}

// FromPascal parses the given PascalCased identifier into its parts.
func FromPascal(ident string) (Parts, error) {
	var (
		elem  []byte
		parts Parts
	)
	// Ensure the first character is upper case.
	if len(ident) > 0 {
		char := ident[0]
		if char < 'A' || char > 'Z' {
			return nil, fmt.Errorf("ident: invalid PascalCased identifier: %q", ident)
		}
		elem = append(elem, char)
	}
	caps := true
	for i := 1; i < len(ident); i++ {
		char := ident[i]
		if char >= 'A' && char <= 'Z' {
			if caps {
				elem = append(elem, char)
			} else {
				caps = true
				parts = parts.add(elem)
				elem = []byte{char}
			}
		} else if caps {
			caps = false
			elem = append(elem, char)
			parts, elem = parts.tryAdd(elem)
		} else {
			elem = append(elem, char)
		}
	}
	if len(elem) > 0 {
		parts = parts.add(elem)
	}
	return parts, nil
}

// FromScreamingSnake parses the given SCREAMING_SNAKE_CASED identifier into its
// parts.
func FromScreamingSnake(ident string) Parts {
	return FromSnake(ident)
}

// FromSnake parses the given snake_cased identifier into its parts.
func FromSnake(ident string) Parts {
	var (
		elem  []byte
		parts Parts
	)
	for i := 0; i < len(ident); i++ {
		char := ident[i]
		if char == '_' {
			if len(elem) == 0 {
				continue
			}
			parts = append(parts, normalize(bytes.ToLower(elem)))
			elem = []byte{}
		} else {
			elem = append(elem, char)
		}
	}
	if len(elem) > 0 {
		parts = append(parts, normalize(bytes.ToLower(elem)))
	}
	return parts
}

func normalize(elem []byte) []byte {
	if special, ok := mapping[string(bytes.ToUpper(elem))]; ok {
		return []byte(special)
	}
	if len(elem) > 0 && 'a' <= elem[0] && elem[0] <= 'z' {
		elem[0] -= 32
	}
	return elem
}
