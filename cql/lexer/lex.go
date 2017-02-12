package lexer

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type ItemType int

func (i ItemType) String() string {
	switch i {
	case ItemError:
		return "ERROR"
	case ItemEOF:
		return "<eof>"
	case ItemKeyword:
		return "KEYWORD"
	case ItemWhitespace:
		return "WS"
	case ItemIdentifier:
		return "IDENTIFIER"
	case ItemInteger:
		return "INTEGER"
	case ItemFloat:
		return "FLOAT"
	case ItemString:
		return "STRING"
	case ItemUUID:
		return "UUID"
	case ItemBoolean:
		return "BOOLEAN"
	case ItemBlob:
		return "BLOB"
	case ItemStar:
		return "STAR"
	case ItemComma:
		return "COMMA"
	case ItemBracket:
		return "BRACKET"
	case ItemSemiColon:
		return "SEMICOLON"
	case ItemDot:
		return "DOT"
	case ItemQuestionMark:
		return "QMARK"
	case ItemColon:
		return "COLON"
	default:
		return fmt.Sprintf("UNKOWN_ITEM_%d", i)
	}
}

const (
	ItemEOF ItemType = iota
	ItemError

	ItemKeyword
	ItemWhitespace
	ItemIdentifier
	ItemStar

	// Constants
	ItemString
	ItemInteger
	ItemFloat
	ItemUUID
	ItemBoolean
	ItemBlob

	ItemComma
	ItemBracket
	ItemSemiColon
	ItemDot
	ItemQuestionMark
	ItemColon
	// TODO: lex comments
)

const eof = 0

type Item struct {
	Typ ItemType
	Val string
}

func (i Item) String() string {
	return fmt.Sprintf("[%v %q]", i.Typ, i.Val)
}

type Lexer struct {
	in    string
	start int
}

func Lex(input string) *Lexer {
	return &Lexer{in: input}
}

func acceptString(in string, strs ...string) bool {
	for _, str := range strs {
		if strings.EqualFold(in, str) {
			return true
		}
	}

	return false
}

func acceptPrefix(in string, strs ...string) bool {
	for _, str := range strs {
		if len(str) > len(in) {
			continue
		}

		got := in[:len(str)]
		if strings.EqualFold(got, str) {
			return true
		}
	}

	return false
}

func isLetter(r rune) bool {
	// cql ident matches [a-zA-Z]
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isHex(r rune) bool {
	switch {
	case '0' <= r && r <= '9':
		return true
	case 'a' <= r && r <= 'f':
		return true
	case 'A' <= r && r <= 'F':
		return true
	default:
		return false
	}
}

func isBlob(s string) bool {
	for _, r := range s {
		if !isHex(r) {
			return false
		}
	}
	return true
}

func isUUID(token string) bool {
	first := true
	pos := 0
	for _, size := range [...]int{8, 4, 4, 4, 12} {
		if !first {
			if token[pos] != '-' {
				return false
			}
			pos++
		} else {
			first = false
		}

		for i := 0; i < size; i++ {
			if !isHex(rune(token[pos+i])) {
				return false
			}
		}
		pos += size
	}

	return true
}

func scanNumber(token string) Item {
	pos := 0
	if token[0] == '-' {
		pos++
	}

	for _, r := range token[pos:] {
		if !isDigit(r) {
			break
		}
		pos++
	}

	typ := ItemInteger
	// TODO: this is messy, improve this to not need len checks everywhere
	if pos < len(token) && token[pos] == '.' {
		typ = ItemFloat
		pos++
		for _, r := range token[pos:] {
			if !isDigit(r) {
				break
			}
			pos++
		}
	}

	if pos == len(token) {
		return Item{typ, token}
	}

	if t := token[pos]; t == 'e' || t == 'E' {
		typ = ItemFloat
		pos++
		// TODO: handle case where we dont lex here and above
		if token[pos] == '-' || token[pos] == '+' {
			pos++
		}
		for _, r := range token[pos:] {
			if !isDigit(r) {
				break
			}
			pos++
		}
	}

	if pos != len(token) {
		return Item{ItemError, token}
	}

	return Item{typ, token}
}

func (l *Lexer) nextToken() string {
	if l.start >= len(l.in) {
		return "" // EOF
	}

	type state int

	const (
		START state = iota
		IN_QUOTE
		IN_IDENT
		IN_SPACE
		IN_NUMBER
	)

	pos := l.start

	var quote rune

	st := START
loop:
	for pos < len(l.in) {
		r, _ := utf8.DecodeRuneInString(l.in[pos:])
		if unicode.IsSpace(r) && st == START && pos > l.start {
			break loop
		}

		switch st {
		case IN_SPACE:
			if !unicode.IsSpace(r) {
				break loop
			}
		case IN_NUMBER:
			switch r {
			case '.', 'e', 'E', '-', '+', 'x', 'X':
			default:
				if !(isDigit(r) || isHex(r)) {
					break loop
				}
			}

		case IN_QUOTE:
			if r == quote {
				if pos+1 < len(l.in) && rune(l.in[pos+1]) == quote {
					pos++
				} else {
					pos++
					break loop
				}
				// TODO: is next ' for escape quote?
			}

		case IN_IDENT:
			switch r {
			case '(', ')', ',', '.':
				break loop
			}

			if unicode.IsSpace(r) {
				break loop
			}

		case START:
			switch r {
			case '(', ')', ',', '.':
				pos++
				break loop
			case '"', '\'':
				quote = r
				st = IN_QUOTE
			case '-':
				st = IN_NUMBER
			default:
				if unicode.IsSpace(r) {
					st = IN_SPACE
				} else if isDigit(r) {
					st = IN_NUMBER
				} else {
					st = IN_IDENT
				}
			}
		}

		pos++
	}

	start := l.start
	l.start = pos
	return l.in[start:pos]
}

func (l *Lexer) Item() Item {
	token := l.nextToken()
	if token == "" {
		return Item{Typ: ItemEOF}
	} else if len(token) == 36 && isUUID(token) {
		return Item{ItemUUID, token}
	} else if acceptString(token, "true", "false") {
		return Item{ItemBoolean, token}
	} else if acceptString(token, "nan", "infinity") {
		return Item{ItemFloat, token}
	} else if acceptPrefix(token, "0x") && isBlob(token[2:]) {
		return Item{ItemBlob, token}
	} else if token == "," {
		return Item{ItemComma, token}
	} else if acceptString(token, "(", ")") {
		return Item{ItemBracket, token}
	} else if token == ";" {
		return Item{ItemSemiColon, token}
	} else if token == "." {
		return Item{ItemDot, token}
	} else if token == "?" {
		return Item{ItemQuestionMark, token}
	} else if token == ":" {
		return Item{ItemColon, token}
	}

	ch, _ := utf8.DecodeRuneInString(token)
	if ch == utf8.RuneError {
		return Item{Typ: ItemEOF}
	} else if unicode.IsSpace(ch) {
		return Item{ItemWhitespace, token}
	} else if isDigit(ch) || ch == '-' {
		return scanNumber(token)
	} else if ch == '\'' {
		return Item{ItemString, token}
	}

	if keywords[strings.ToUpper(token)] {
		return Item{ItemKeyword, strings.ToLower(token)}
	}

	if isLetter(ch) || ch == '"' {
		return Item{ItemIdentifier, token}
	}

	return Item{ItemError, token}
}

func (l *Lexer) Tokens() []Item {
	var items []Item

	for {
		item := l.Item()

		items = append(items, item)
		if item.Typ == ItemEOF {
			return items
		}
	}
}

func (l *Lexer) ItemNoWS() Item {
	for {
		item := l.Item()
		if item.Typ != ItemWhitespace {
			return item
		}
	}
}
