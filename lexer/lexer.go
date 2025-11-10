package lexer

import (
	"strings"
	"unicode"
)

type TokenType string

const (
	Ident   TokenType = "IDENT"
	Number  TokenType = "NUMBER"
	Symbol  TokenType = "SYMBOL"
	Keyword TokenType = "KEYWORD"
)

type Token struct {
	Type  TokenType
	Value string
}

var keywords = map[string]bool{
	"fn": true, "let": true, "mut": true, "return": true, "if": true, "else": true,
}

func Lex(src string) []Token {
	var tokens []Token
	var current strings.Builder

	flush := func(tt TokenType) {
		if current.Len() == 0 {
			return
		}
		value := current.String()
		if tt == Ident && keywords[value] {
			tokens = append(tokens, Token{Keyword, value})
		} else {
			tokens = append(tokens, Token{tt, value})
		}
		current.Reset()
	}

	for _, ch := range src {
		switch {
		case unicode.IsSpace(ch):
			flush(Ident)
		case unicode.IsLetter(ch):
			current.WriteRune(ch)
		case unicode.IsDigit(ch):
			current.WriteRune(ch)
		default:
			flush(Ident)
			tokens = append(tokens, Token{Symbol, string(ch)})
		}
	}
	flush(Ident)
	return tokens
}
