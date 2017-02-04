package repl

import (
	"reflect"
	"sort"
	"testing"
)

func TestCommonPrefixLen(t *testing.T) {
	tests := [...]struct {
		a, b string
		plen int
	}{
		{"", "", 0},
		{"a", "b", 0},
		{"a", "a", 1},
		{"aaa", "aab", 2},
		{"a", "aab", 1},
	}

	for _, test := range tests {
		t.Run(test.a+"/"+test.b, func(t *testing.T) {
			pl := commonPrefixLen(test.a, test.b)
			if test.plen != pl {
				t.Fatalf("expeceted %d got %d", test.plen, pl)
			}
		})
	}
}

func TestPrefixCompleter_Insert(t *testing.T) {
	var p trieNode
	p.insert("test")
	if len(p.children) != 1 {
		t.Fatalf("did not insert into empty trie: %v", p)
	} else if p.children[0].prefix != "test" {
		t.Fatalf("expected prefix to be %q got %q", "test", p.children[0].prefix)
	}

	p.insert("horse")
	if len(p.children) != 2 {
		t.Fatalf("did not insert into trie: %v", p)
	} else if p.children[1].prefix != "horse" {
		t.Fatalf("expected prefix to be %q got %q", "horse", p.children[1].prefix)
	}

	p.insert("horses")
	if len(p.children) != 2 {
		t.Fatalf("did not insert into root of trie: %v", p)
	} else if p.children[1].prefix != "horse" {
		t.Fatalf("expected prefix to be %q got %q", "horse", p.children[1].prefix)
	} else {
		child := p.children[1]
		if len(child.children) != 2 {
			t.Fatalf("expected 2 children got %d", len(child.children))
		}

		for _, node := range child.children {
			if node.prefix != terminal {
				child = node
				break
			}
		}

		if len(child.children) != 1 {
			t.Fatalf("expected 1 child got %d", len(child.children))
		}

	}

	p.insert("house")
	if len(p.children) != 2 {
		t.Fatalf("did not insert into trie: %v", p)
	} else if p.children[1].prefix != "ho" {
		t.Fatalf("expected prefix to be %q got %q", "ho", p.children[1].prefix)
	}
}

func TestPrefixCompleter_ContainsInsert(t *testing.T) {
	values := []string{"house", "horse", "horses", "him", "his", "her", "potato", "pot", "plant", "nope"}
	sort.Strings(values)

	var p trieNode
	for _, v := range values {
		p.insert(v)
	}

	for _, v := range values {
		if !p.contains(v) {
			t.Errorf("did not contain %q", v)
		}
	}
}

func TestPrefixCompleter_All(t *testing.T) {
	values := []string{"house", "horse", "horses", "him", "his", "her", "potato", "pot", "plant", "nope", "a", "a_c", "a_b"}
	sort.Strings(values)

	var p trieNode
	for _, v := range values {
		p.insert(v)
	}

	all := p.All()
	sort.Strings(all)

	if !reflect.DeepEqual(all, values) {
		t.Fatalf("expected %q got %q", values, all)
	}
}

func TestComplete(t *testing.T) {
	tests := []struct {
		item         string
		trieContents []string
		result       []string
	}{
		{
			"h",
			[]string{"horse"},
			[]string{"orse"},
		},
		{
			"missing",
			[]string{"horse"},
			nil,
		},
		{
			"ho",
			[]string{"house", "horse"},
			[]string{"rse", "use"},
		},
		{
			"ho",
			[]string{"house", "test"},
			[]string{"use"},
		},
		{
			"horses",
			[]string{"house", "horse", "horses"},
			[]string{""},
		},
		{
			"sys",
			[]string{"system_keyspaces", "system_tables", "system"},
			[]string{"tem_keyspaces", "tem_tables", "tem"},
		},
	}

	for _, test := range tests {
		sort.Strings(test.result)
		t.Run(test.item, func(t *testing.T) {
			var p trieNode
			for _, s := range test.trieContents {
				p.insert(s)
			}

			result := p.Complete(test.item)
			sort.Strings(result)
			if !reflect.DeepEqual(result, test.result) {
				t.Fatalf("expected predictions %q got %q", test.result, result)
			}
		})
	}
}

func TestPrefixCompleter_Contains_Missing(t *testing.T) {
	var p trieNode
	p.insert("a")
	p.insert("a_b")
	p.insert("a_c")
	if p.contains("a_") {
		t.Fatalf("should not contain %q", "a_")
	}
}

func TestPrefixCompleter_CorrectInsert(t *testing.T) {
	var p trieNode
	p.insert("a_b")
	p.insert("a_c")
	p.insert("a")

	if len(p.children) != 1 {
		t.Fatalf("expected single child got %d", len(p.children))
	}

	child := p.children[0]
	if child.prefix != "a" {
		t.Fatalf("expected first node to be %q got %q", "a", child.prefix)
	} else if len(child.children) != 2 {
		t.Fatalf("a should have 2 child nodes got %d", len(child.children))
	}

	for _, node := range child.children {
		if node.prefix != terminal {
			child = node
			break
		}
	}

	if child.prefix != "_" {
		t.Fatalf("child should have prefix %q got %q", "_", child.prefix)
	}
}
