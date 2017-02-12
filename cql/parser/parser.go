package parser

import "github.com/gocql/gocqlsh/cql/lexer"

type Statement interface {
}

func Parse(lex *lexer.Lexer) (Statement, error) {
	keyword := lex.ItemNoWS()

	switch keyword.Val {
	case "insert":
		return parseInsert(lex)
	}

	panic(keyword)
}
