package token

import (
	"strconv"
)

// Token Javascript中的词标集合
type Token int

func (t Token) String() string {
	if t == 0 {
		return "UNKNOWN"
	}
	if t < Token(len(token2string)) {
		return token2string[t]
	}
	return "token(" + strconv.Itoa(int(t)) + ")"
}

func (t Token) precedence(in bool) int {
	switch t {
	case LOGICAL_OR:
		return 1
	case LOGICAL_AND:
		return 2
	case OR, OR_ASSIGN:
		return 3
	case EXCLUSIVE_OR:
		return 4
	case AND, AND_ASSIGN:
		return 5
	case EQUAL, NOT_EQUAL, STRICT_EQUAL, STRICT_NOT_EQUAL:
		return 6
	case LESS, GREATER, LESS_OR_EQUAL, GREATER_OR_EQUAL, INSTANCEOF:
		return 7
	case IN:
		if in {
			return 7
		}
		return 0
	case SHIFT_LEFT, SHIFT_RIGHT, UNSIGNED_SHIFT_RIGHT:
		fallthrough
	case SHIFT_LEFT_ASSIGN, SHIFT_RIGHT_ASSIGN, UNSIGNED_SHIFT_RIGHT_ASSIGN:
		return 8
	case PLUS, MINUS, ADD_ASSIGN, SUBTRACT_ASSIGN:
		return 9
	case MULTIPLY, SLASH, REMAINDER, MULTIPLY_ASSIGN, QUOTIENT_ASSIGN, REMAINDER_ASSIGN:
		return 11
	}
	return 0
}

type keyword struct {
	token         Token
	futureKeyword bool
	strict        bool
}

// IsKeyword 如果 literal 是一个关键词，则返回一个关键字令牌
// 如果 literal 是一个未来的关键字，或者不是关键字，则返回0
func IsKeyword(literal string) (Token, bool) {
	if kw, exists := keywordTable[literal]; exists {
		if kw.futureKeyword {
			return KEYWORD, kw.strict
		}
		return kw.token, false
	}
	return 0, false
}

func IsId(t Token) bool {
	return t >= IDENTIFIER
}

func IsUnreservedWord(t Token) bool {
	return t > ESCAPED_RESERVED_WORD
}
