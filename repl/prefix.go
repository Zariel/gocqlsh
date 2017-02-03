package repl

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

// all terms are inserted as term$ to indicate the full term
const terminal = "$"

type trieNode struct {
	prefix   string
	children []*trieNode
}

func (p *trieNode) insert(item string) {
	if p.prefix == terminal {
		panic("can not insert value into terminal")
	} else if item == "" {
		panic("can not insert empty term")
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

func (p trieNode) all(prefix string, res []string) []string {
	for _, c := range p.children {
		if c.prefix == terminal {
			// prefix is a complete term
			res = append(res, prefix)
		} else {
			res = c.all(prefix+c.prefix, res)
		}
	}

	return res
}

func (p trieNode) All() []string {
	return p.all("", nil)
}

func (p trieNode) complete(prefix, term string) []string {
	// empty term, we have recursed down the trie a bit and run out of term, all sub terms are possible from here
	if term == "" {
		return p.all(prefix, nil)
	}

	for _, node := range p.children {
		plen := commonPrefixLen(term, node.prefix)
		if plen > 0 {
			return node.complete(prefix+node.prefix, term[plen:])
		}
	}

	return nil
}

func (p trieNode) Complete(term string) []string {
	// TODO: this is silly, do this in 1 pass no need to strip the prefix of the results
	res := p.complete(p.prefix, term)
	for i, complete := range res {
		res[i] = complete[commonPrefixLen(term, complete):]
	}
	return res
}

func prefixComplete(term string, items ...string) []string {
	// TODO: this needs to be able to handle case insensitive matching, ie, given IN return INSERT or insert
	var p trieNode
	for _, item := range items {
		p.insert(item)
	}
	return p.Complete(term)
}
