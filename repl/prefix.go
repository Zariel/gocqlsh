package repl

import "log"

func commonPrefixLen(a, b string) int {
	n := len(a)
	if len(b) < n {
		n = len(b)
	}

	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			return i
		}
	}

	return n
}

type trieNode struct {
	prefix   string
	children []*trieNode
}

func (p *trieNode) insert(item string) {
	if p.prefix == terminal {
		panic("can not insert value into terminal")
	}

	if item == terminal {
		// check if we already have this term
		for _, node := range p.children {
			if node.prefix == terminal {
				return
			}
		}

		node := &trieNode{prefix: item}
		p.children = append(p.children, node)
		if p.prefix == terminal {
			panic("can not insert terminal into a terminal")
		}
		return
	}

	// TODO: keep the list of nodes sorted by prefix to enable binary search
	for _, node := range p.children {
		// TODO: handle split
		plen := commonPrefixLen(node.prefix, item)
		if plen > 0 && node.prefix != terminal {
			if plen == len(item) {
				// overlap, insert a terminal
				node.insert(terminal)
			} else if plen == len(node.prefix) {
				// log.Printf("overlapping item=%q node=%q", item, node.prefix)
				// have some overlap by len(item) > len(node.prefix)
				node.insert(item[plen:])
			} else {
				// node is a full overlap, need to reshuffle the tree
				prefix := node.prefix[:plen]
				suffix := node.prefix[plen:]
				node.prefix = prefix

				children := node.children
				newNode := &trieNode{prefix: suffix, children: children}

				node.children = []*trieNode{newNode}
				node.insert(item[plen:])
			}

			return
		}
	}

	node := &trieNode{prefix: item}
	p.children = append(p.children, node)
	node.insert(terminal)
}

func (p trieNode) contains(item string) bool {
	// TODO: move this to a root trie node structure, then just
	// search for a suffix
	for _, c := range p.children {
		if c.prefix == terminal {
			return true
		}

		prefix := commonPrefixLen(c.prefix, item)
		if prefix > 0 {
			return c.contains(item[prefix:])
		}
	}

	return false
}

func (p trieNode) all(partial string) []string {
	// FIME
	var res []string
	for _, c := range p.children {
		if c.prefix == terminal {
			res = append(res, partial)
		} else {
			res = append(res, c.all(partial+c.prefix)...)
		}
	}

	return res
}

func (p trieNode) complete(prefix, s string) []string {
	if s == "" {
		return p.all(prefix)
	}

	for _, node := range p.children {
		plen := commonPrefixLen(s, node.prefix)
		if plen > 0 {
			log.Println(node)
		}
	}
	return nil
}

func (p trieNode) Complete(s string) []string {
	for _, node := range p.children {
		plen := commonPrefixLen(s, node.prefix)
		if plen > 0 {
			return p.complete("", s[plen:])
		}
	}

	return nil
}

const terminal = "$"

func prefixComplete(items ...string) trieNode {
	var p trieNode
	for _, item := range items {
		log.Println("inserting:", item)
		p.insert(item)
	}
	return p
}
