package sql

import (
	"fmt"
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokEOF tokenKind = iota
	tokIdent
	tokNumber
	tokString
	tokComma
	tokLParen
	tokRParen
	tokStar
	tokOp
	tokKeyword
)

type token struct {
	kind tokenKind
	val  string
}

var keywords = map[string]bool{
	"SELECT": true, "FROM": true, "WHERE": true, "GROUP": true, "BY": true,
	"ORDER": true, "LIMIT": true, "ASC": true, "DESC": true,
	"AND": true, "OR": true, "NOT": true, "AS": true,
	"COUNT": true, "SUM": true, "AVG": true, "MIN": true, "MAX": true,
	"LIKE": true, "IN": true, "IS": true, "NULL": true,
}

func tokenize(s string) ([]token, error) {
	var toks []token
	i := 0
	for i < len(s) {
		c := s[i]
		if unicode.IsSpace(rune(c)) {
			i++
			continue
		}
		if c == ',' {
			toks = append(toks, token{kind: tokComma, val: ","})
			i++
			continue
		}
		if c == '(' {
			toks = append(toks, token{kind: tokLParen, val: "("})
			i++
			continue
		}
		if c == ')' {
			toks = append(toks, token{kind: tokRParen, val: ")"})
			i++
			continue
		}
		if c == '*' {
			toks = append(toks, token{kind: tokStar, val: "*"})
			i++
			continue
		}
		if c == '\'' || c == '"' {
			end := strings.IndexByte(s[i+1:], c)
			if end < 0 {
				return nil, fmt.Errorf("unterminated string literal")
			}
			toks = append(toks, token{kind: tokString, val: s[i+1 : i+1+end]})
			i += end + 2
			continue
		}
		if c == '<' || c == '>' || c == '=' || c == '!' {
			op := string(c)
			if i+1 < len(s) && (s[i+1] == '=' || (c == '<' && s[i+1] == '>')) {
				op += string(s[i+1])
				i += 2
			} else {
				i++
			}
			toks = append(toks, token{kind: tokOp, val: op})
			continue
		}
		if unicode.IsDigit(rune(c)) || (c == '-' && i+1 < len(s) && unicode.IsDigit(rune(s[i+1]))) {
			start := i
			if c == '-' {
				i++
			}
			for i < len(s) && (unicode.IsDigit(rune(s[i])) || s[i] == '.') {
				i++
			}
			toks = append(toks, token{kind: tokNumber, val: s[start:i]})
			continue
		}
		if unicode.IsLetter(rune(c)) || c == '_' {
			start := i
			for i < len(s) && (unicode.IsLetter(rune(s[i])) || unicode.IsDigit(rune(s[i])) || s[i] == '_' || s[i] == '.') {
				i++
			}
			word := s[start:i]
			upper := strings.ToUpper(word)
			if keywords[upper] {
				toks = append(toks, token{kind: tokKeyword, val: upper})
			} else {
				toks = append(toks, token{kind: tokIdent, val: word})
			}
			continue
		}
		return nil, fmt.Errorf("unexpected character %q at position %d", c, i)
	}
	toks = append(toks, token{kind: tokEOF})
	return toks, nil
}
