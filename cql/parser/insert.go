package parser

import (
	"errors"
	"fmt"
	"log"

	"github.com/gocql/gocqlsh/cql/lexer"
)

type node struct {
	item lexer.Item
	prev *node
}

type InsertStatement struct {
	Keyspace string
	Table    string
	Columns  []string
	Values   []lexer.Item
	JSON     bool

	// nodes is a reverse linked list of the items which hav been lexed
	nodes *node
}

func (s *InsertStatement) String() string {
	return fmt.Sprintf("[insert keyspace=%q table=%q columns=%q values=%v json=%v]",
		s.Keyspace, s.Table, s.Columns, s.Values, s.JSON)
}

func parseInsert(lex *lexer.Lexer) (*InsertStatement, error) {
	const (
		into = iota
		tableName
		keyspace
		maybeDot
		table
		maybeJson
		jsonKeyword
		jsonString
		nameList
		moreColumns
		column
		valuesKeyword
		namelistOpen
		namelistClose
		valuesListOpen
		valuesListClose
		valuesList
		moreValues
		value
		ifNotUsing
		ifNotExists
		using
	)
	state := into

	var (
		insert InsertStatement
		last   lexer.Item
	)

loop:
	for {
		// TODO: should we be able to just unwind the lexer?
		item := lex.ItemNoWS()
		insert.nodes = &node{item: item, prev: insert.nodes}
		switch item.Typ {
		case lexer.ItemEOF:
			break loop
		case lexer.ItemError:
			// what do?
			return nil, errors.New(item.Val)
		}

		// this will be _much_ nicer as a functional state machine, eg
		// return parseNameList() // bracket, parseColumn, for _, parse column
	stateSwitch:
		log.Println(item, state)
		switch state {
		case into:
			if item != (lexer.Item{lexer.ItemKeyword, "into"}) {
				break loop
			}
			state = tableName

		case tableName:
			if item.Typ != lexer.ItemIdentifier {
				break loop
			}
			last = item
			state = maybeDot

		case maybeDot:
			// here should we just unwind the lexer and say, state = maybeJson?
			switch item.Typ {
			case lexer.ItemDot:
				insert.Keyspace = last.Val
				state = table
			case lexer.ItemKeyword:
				insert.Table = last.Val
				state = jsonKeyword
				goto stateSwitch
			case lexer.ItemBracket:
				insert.Table = last.Val
				state = namelistOpen
				goto stateSwitch
			default:
				break loop
			}

		case namelistOpen:
			if item != (lexer.Item{lexer.ItemBracket, "("}) {
				break loop
			}
			state = nameList

		case jsonKeyword:
			if item.Val != "JSON" {
				break loop
			}
			insert.JSON = true
			state = jsonString

		case table:
			if item.Typ != lexer.ItemIdentifier {
				break loop
			}

			insert.Table = item.Val
			state = namelistOpen

		case nameList:
			if item.Typ != lexer.ItemIdentifier {
				break loop
			}

			insert.Columns = append(insert.Columns, item.Val)
			state = moreColumns

		case moreColumns:
			switch item.Typ {
			case lexer.ItemComma:
				state = column
			case lexer.ItemBracket:
				state = namelistClose
				goto stateSwitch
			default:
				break loop
			}

		case column:
			if item.Typ != lexer.ItemIdentifier {
				break loop
			}

			insert.Columns = append(insert.Columns, item.Val)
			state = moreColumns

		case namelistClose:
			if item != (lexer.Item{lexer.ItemBracket, ")"}) {
				break loop
			}

			state = valuesKeyword

		case valuesKeyword:
			if item != (lexer.Item{lexer.ItemKeyword, "values"}) {
				break loop
			}

			state = valuesListOpen

		case valuesListOpen:
			if item != (lexer.Item{lexer.ItemBracket, "("}) {
				break loop
			}

			state = valuesList

		case valuesList:
			switch item.Typ {
			// term
			case lexer.ItemInteger, lexer.ItemFloat, lexer.ItemString, lexer.ItemBoolean, lexer.ItemBlob, lexer.ItemUUID:
				insert.Values = append(insert.Values, item)
				state = moreValues
			case lexer.ItemQuestionMark:
				insert.Values = append(insert.Values, item)
				state = moreValues
				// TOOD: collection
				// TODO: variable
				// TODO: function
			default:
				break loop
			}

		case moreValues:
			switch item.Typ {
			case lexer.ItemComma:
				state = valuesList
			case lexer.ItemBracket:
				state = valuesListClose
				goto stateSwitch
			default:
				break loop
			}

		case valuesListClose:
			if item != (lexer.Item{lexer.ItemBracket, "("}) {
				break loop
			}

			state = ifNotUsing

		case ifNotUsing:
			switch item {
			case lexer.Item{lexer.ItemKeyword, "IF"}:
				state = ifNotExists
				goto stateSwitch
			case lexer.Item{lexer.ItemKeyword, "USING"}:
				state = using
				goto stateSwitch
			default:
				break loop
			}

		default:
			panic(fmt.Sprintf("unknown state %d", state))
		}

		last = item
	}
	log.Println("end", last, state)

	return &insert, nil
}
