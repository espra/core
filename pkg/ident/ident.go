// Public Domain (-) 2018-present, The Core Authors.
// See the Core UNLICENSE file for details.

// Package ident provides support for converting identifiers between different
// case styles.
package ident

import (
	"bytes"
)

// Parts represents the normalised elements of an identifier.
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
	last := len(p) - 1
	for idx, part := range p {
		out = append(out, bytes.ToLower(part)...)
		if idx != last {
			out = append(out, '-')
		}
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
	last := len(p) - 1
	for idx, part := range p {
		out = append(out, bytes.ToUpper(part)...)
		if idx != last {
			out = append(out, '_')
		}
	}
	return string(out)
}

// ToSnake converts the identifier into a snake_cased string.
func (p Parts) ToSnake() string {
	var out []byte
	last := len(p) - 1
	for idx, part := range p {
		out = append(out, bytes.ToLower(part)...)
		if idx != last {
			out = append(out, '_')
		}
	}
	return string(out)
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
	return append(parts, FromPascal(ident[i:])...)
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
func FromPascal(ident string) Parts {
	var (
		elem  []byte
		parts Parts
	)
	caps := true
	for i := 0; i < len(ident); i++ {
		char := ident[i]
		if char >= 'A' && char <= 'Z' {
			if caps {
				elem = append(elem, char)
			} else {
				caps = true
				parts = processLeftover(parts, elem)
				elem = []byte{char}
			}
		} else if caps {
			caps = false
			elem = append(elem, char)
			parts, elem = process(parts, elem)
		} else {
			elem = append(elem, char)
		}
	}
	if len(elem) > 0 {
		parts = processLeftover(parts, elem)
	}
	return parts
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

func process(parts Parts, elem []byte) (Parts, []byte) {
	var nelem []byte
	// Try to match exactly.
	if special, ok := mapping[string(bytes.ToUpper(elem))]; ok {
		return append(parts, []byte(special)), nil
	}
	// Try to match from the end for the longest identifier with a non-uppercase
	// suffix.
	last := ""
	pos := -1
	for i := len(elem) - 1; i >= 0; i-- {
		if special, ok := mapping[string(bytes.ToUpper(elem[i:]))]; ok {
			last = special
			pos = i
		}
	}
	if pos == -1 {
		nelem = elem[len(elem)-2:]
		elem = elem[:len(elem)-2]
	} else {
		elem = elem[:pos]
	}
	// Try to find the longest matches from the start.
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
			parts = append(parts, elem)
			break
		}
		parts = append(parts, []byte(match))
		elem = elem[pos:]
	}
	if len(last) > 0 {
		parts = append(parts, []byte(last))
	}
	return parts, nelem
}

func processLeftover(parts Parts, elem []byte) Parts {
	// Try to match exactly.
	if special, ok := mapping[string(bytes.ToUpper(elem))]; ok {
		return append(parts, []byte(special))
	}
	// Try to find the longest matches from the start.
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
			parts = append(parts, elem)
			break
		}
		parts = append(parts, []byte(match))
		elem = elem[pos:]
	}
	return parts
}
