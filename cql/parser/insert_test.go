package parser

import (
	"testing"

	"github.com/gocql/gocqlsh/cql/lexer"
)

func TestParseInsert(t *testing.T) {
	insert, err := parseInsert(lexer.Lex(" INTO someKeyspace.someTable(col1, col2) VALUES('potato', ?);"))
	if err != nil {
		t.Fatal(err)
	}

	t.Fatal(insert)
}
