package lexer

import (
	"reflect"
	"testing"
)

func TestIsDigit(t *testing.T) {
	const s = "0123456789"
	for _, r := range s {
		if !isDigit(r) {
			t.Errorf("%c should be a digit", r)
		}
	}
}

func TestIsLetter(t *testing.T) {
	const s = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for _, r := range s {
		if !isLetter(r) {
			t.Errorf("%c should be a letter", r)
		}
	}
}

func TestLex(t *testing.T) {
	tests := [...]struct {
		typ    ItemType
		inputs []string
	}{
		{
			ItemInteger,
			[]string{"0", "-1", "12309124801293"},
		},
		{
			ItemFloat,
			[]string{"1.0", "-1.0", "31.231E-1212", "31.231e+1212", "nan", "INFINITY"},
		},
		{
			ItemBoolean,
			[]string{"true", "false", "TRUE", "FALSE"},
		},
		{
			ItemString,
			[]string{"'raw string'", "'escaped ''string'"},
		},
		{
			ItemUUID,
			[]string{"f2c993b9-6c2f-4137-a8f0-0fd5e5cc4433", "F2C993B9-6C2F-4137-A8F0-0FD5E5CC4433"},
		},
		{
			ItemBlob,
			[]string{"0xFF", "0X23eFaa"},
		},
		{
			ItemIdentifier,
			[]string{"sometable", "INFINITYtable", `"quoted ident"`, `"nested "" quote"`},
		},
		{
			ItemKeyword,
			[]string{"select", "into", "insert", "update", "create"},
		},
		{
			ItemComma,
			[]string{","},
		},
		{
			ItemSemiColon,
			[]string{";"},
		},
	}

	for _, test := range tests {
		t.Run(test.typ.String(), func(t *testing.T) {
			for _, in := range test.inputs {
				t.Run(in, func(t *testing.T) {
					item := Lex(in).Item()
					if item.Typ != test.typ {
						t.Fatalf("expected to get %v got %v", test.typ, item)
					} else if item.Val != in {
						t.Fatalf("expected to get %q got %q", in, item.Val)
					}
				})
			}
		})
	}
}

func TestLexScanToken(t *testing.T) {
	tests := [...]struct {
		in  string
		exp []string
	}{
		{"test simple input", []string{"test", " ", "simple", " ", "input", ""}},
		{"single", []string{"single", ""}},
		{" ", []string{" ", ""}},
		{`  "quoted"`, []string{"  ", `"quoted"`, ""}},
		{"table(column, col2)", []string{"table", "(", "column", ",", " ", "col2", ")", ""}},
	}

	for _, test := range tests {
		t.Run(test.in, func(t *testing.T) {
			l := Lex(test.in)
			var tokens []string

			for {
				token := l.nextToken()
				tokens = append(tokens, token)
				if token == "" {
					break
				}
			}

			if !reflect.DeepEqual(test.exp, tokens) {
				t.Fatalf("expected %q got %q", test.exp, tokens)
			}
		})
	}
}
