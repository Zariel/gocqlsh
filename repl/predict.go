package repl

import (
	"bytes"
	"log"

	"github.com/gocql/gocql"
	"github.com/gocql/gocqlsh/cql/lexer"

	"github.com/chzyer/readline"
)

type cqlCompleter struct {
	db *gocql.Session
}

func (c *cqlCompleter) Print(prefix string, level int, buf *bytes.Buffer) {
	panic("nope")
}

func (c *cqlCompleter) Do(line []rune, pos int) (newLine [][]rune, offset int) {
	if pos != len(line) {
		// TODO; handle middle of word prediction
		return nil, 0
	}

	lines := c.queryParser(string(line))
	runes := make([][]rune, len(lines))
	for i, line := range lines {
		runes[i] = []rune(line)
	}

	return runes, 0
}

func (c *cqlCompleter) GetName() []rune {
	return []rune("potato")
}

func (c *cqlCompleter) GetChildren() []readline.PrefixCompleterInterface {
	return nil
}

func (c *cqlCompleter) SetChildren(children []readline.PrefixCompleterInterface) {
}

type completer struct {
	l     *lexer.Lexer
	last  lexer.Item
	items []string
}

func (c *completer) Next() bool {
	if len(c.items) > 0 {
		return false
	}

	c.last = c.l.Item()
	if c.last.Typ == lexer.ItemEOF {
		return false
	}
	return true
}

func (c *completer) Space() {
	if len(c.items) > 0 {
		return
	}

	c.Next()

	if c.last.Typ != lexer.ItemWhitespace {
		c.items = append(c.items, " ")
	}
}

func (c *completer) Expect(token string) {
	if len(c.items) > 0 {
		return
	}

	c.Next()

	if c.last.Val != token {
		c.items = append(c.items, token[commonPrefixLen(token, c.last.Val):])
	}
}

func (c *completer) Accept(typ lexer.ItemType, fn func() []string) string {
	if len(c.items) > 0 {
		return ""
	}

	c.Next()
	if c.last.Typ != typ {
		c.items = append(c.items, fn()...)
		return ""
	} else {
		return c.last.Val
	}
}

func (c *cqlCompleter) completeInsert(lex *lexer.Lexer) []string {
	comp := &completer{l: lex}

	comp.Space()
	comp.Expect("into")
	comp.Space()

	// TODO: this could be table or keyspace, need to figure out
	keyspace := comp.Accept(lexer.ItemIdentifier, func() []string {
		// TODO: move this to gocql
		var keyspaces []string

		s := c.db.Query("SELECT keyspace_name FROM system.schema_keyspaces").Iter().Scanner()
		for s.Next() {
			var name string
			if err := s.Scan(&name); err != nil {
				log.Println(err)
				return nil
			}

			keyspaces = append(keyspaces, name)
		}

		if err := s.Err(); err != nil {
			log.Println(err)
		}

		return keyspaces
	})

	if keyspace == "" {
		return comp.items
	}

	keyspaceMeta, err := c.db.KeyspaceMetadata(keyspace)
	if err != nil {
		// TODO: need to output errors somewhere
		log.Println(err)
		return comp.items
	}

	comp.Expect(".")
	table := comp.Accept(lexer.ItemIdentifier, func() []string {
		var tables []string
		for table := range keyspaceMeta.Tables {
			tables = append(tables, table)
		}
		return tables
	})

	if table == "" {
		return comp.items
	}

	comp.Expect("(")

	var columns []string

	// column list
	for comp.Next() {
		// could do some better type checking here, ie col [, col]*
		switch comp.last.Typ {
		case lexer.ItemComma:
			comp.Space()
			col := comp.Accept(lexer.ItemIdentifier, func() []string {
				m, ok := keyspaceMeta.Tables[table]
				if !ok {
					return nil
				}

				return m.OrderedColumns
			})

			columns = append(columns, col)
		case lexer.ItemIdentifier:
			comp.Expect(",")
			// TODO: be nice to return ", " here
			comp.Space()
		case lexer.ItemBracket:
			break
		}
	}

	return comp.items
}

func (c *cqlCompleter) queryParser(q string) []string {
	l := lexer.Lex(q)
	keyword := l.ItemNoWS()
	if keyword.Typ != lexer.ItemKeyword {
		log.Printf("%v", keyword)
		return nil
	}

	switch keyword.Val {
	case "select":
		return nil
	case "insert":
		return c.completeInsert(l)
	}

	return prefixComplete(keyword.Val, "insert", "select", "update", "delete")
}
