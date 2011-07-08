// Public Domain (-) 2011 The Ampify Authors.
// See the Ampify UNLICENSE file for details.

package structure

import (
	"bytes"
	"fmt"
)

type PrefixTreeNode struct {
	label string
	nodes []*PrefixTreeNode
	value interface{}
}

type PrefixTree struct {
	root *PrefixTreeNode
	size int
}

type PrefixMatch struct {
	Suffix string
	Value  interface{}
}

func (tree *PrefixTree) Insert(key string, value interface{}) {
	if len(key) == 0 {
		return
	}
	char := key[0]
	node := tree.root
	for _, subnode := range tree.root.nodes {
		if subnode.label[0] == char {
			node = subnode
		outer:
			for {
				if key == node.label {
					if node.value == nil {
						tree.size += 1
					}
					node.value = value
					return
				}
				match := 0
				max := len(node.label)
				for i, b := range []byte(key) {
					if i == max {
						break
					}
					if b == node.label[i] {
						match = i + 1
					} else {
						break
					}
				}
				key = key[match:]
				if match < max {
					nodes := make([]*PrefixTreeNode, 1)
					nodes[0] = &PrefixTreeNode{
						label: node.label[match:],
						nodes: node.nodes,
						value: node.value,
					}
					node.label = node.label[:match]
					node.nodes = nodes
					node.value = nil
					break
				}
				char = key[0]
				for _, subnode := range node.nodes {
					if subnode.label[0] == char {
						node = subnode
						continue outer
					}
				}
				break
			}
		}
	}
	tree.size += 1
	nodes := make([]*PrefixTreeNode, 0)
	node.nodes = append(
		node.nodes,
		&PrefixTreeNode{nodes: nodes, label: key, value: value},
	)
}

func (tree *PrefixTree) Lookup(key string) (value interface{}) {
	if len(key) == 0 {
		return
	}
	char := key[0]
	for _, node := range tree.root.nodes {
		if node.label[0] == char {
		outer:
			for {
				if key == node.label {
					return node.value
				}
				max := len(node.label)
				if max > len(key) || key[:max] != node.label {
					return
				}
				key = key[max:]
				char = key[0]
				for _, subnode := range node.nodes {
					if subnode.label[0] == char {
						node = subnode
						continue outer
					}
				}
				return
			}
			break
		}
	}
	return
}

func (tree *PrefixTree) Delete(key string) {
	if len(key) == 0 {
		return
	}
	char := key[0]
	nodes := make([]*PrefixTreeNode, 0)
	for _, node := range tree.root.nodes {
		if node.label[0] == char {
		outer:
			for {
				if key == node.label {
					if node.value == nil {
						return
					}
					node.value = nil
					tree.size -= 1
					if len(node.nodes) != 0 {
						return
					}
					nodes = append(nodes, node)
					break
				}
				max := len(node.label)
				if max > len(key) || key[:max] != node.label {
					return
				}
				key = key[max:]
				char = key[0]
				for _, subnode := range node.nodes {
					if subnode.label[0] == char {
						nodes = append(nodes, node)
						node = subnode
						continue outer
					}
				}
				return
			}
			break
		}
	}
	var node, prev *PrefixTreeNode
	for i := len(nodes) - 1; i >= 0; i-- {
		node = nodes[i]
		if i != 0 {
			prev = nodes[i-1]
		} else {
			prev = tree.root
		}
		if node.value == nil {
			if len(prev.nodes) == 1 {
				if prev == tree.root {
					prev.nodes = prev.nodes[:0]
				} else {
					continue
				}
			} else {
				clone := make([]*PrefixTreeNode, len(prev.nodes)-1)
				j := 0
				for _, n := range prev.nodes {
					if n != node {
						clone[j] = n
						j += 1
					}
				}
				prev.nodes = clone
				break
			}
		}
	}
}

func (tree *PrefixTree) MatchPrefix(key string) (match []*PrefixMatch) {
	if len(key) == 0 {
		return
	}
	match = make([]*PrefixMatch, 0)
	char := key[0]
	for _, node := range tree.root.nodes {
		if node.label[0] == char {
		outer:
			for {
				if key == node.label {
					if node.value == nil {
						return
					}
					match = append(match, &PrefixMatch{
						Suffix: "",
						Value:  node.value,
					})
					return
				}
				max := len(node.label)
				if max > len(key) || key[:max] != node.label {
					return
				}
				key = key[max:]
				char = key[0]
				if node.value != nil {
					match = append(match, &PrefixMatch{
						Suffix: key,
						Value:  node.value,
					})
				}
				for _, subnode := range node.nodes {
					if subnode.label[0] == char {
						node = subnode
						continue outer
					}
				}
				return
			}
			break
		}
	}
	return
}

func (tree *PrefixTree) Size() int {
	return tree.size
}

func (tree *PrefixTree) String() string {
	buffer := &bytes.Buffer{}
	fmt.Fprintf(buffer, "# PrefixTree: %d\n", tree.Size())
	for _, node := range tree.root.nodes {
		node.dump(buffer, "  ")
	}
	return buffer.String()
}

func (node *PrefixTreeNode) dump(buffer *bytes.Buffer, indent string) {
	if node.value == nil {
		fmt.Fprintf(buffer, "%s- %s:\n", indent, node.label)
	} else {
		fmt.Fprintf(buffer, "%s- %s: %v\n", indent, node.label, node.value)
	}
	if len(node.nodes) != 0 {
		indent += "  "
		for _, subnode := range node.nodes {
			subnode.dump(buffer, indent)
		}
	}
}

func NewPrefixTree() *PrefixTree {
	nodes := make([]*PrefixTreeNode, 0)
	tree := &PrefixTree{
		root: &PrefixTreeNode{nodes: nodes},
	}
	return tree
}
