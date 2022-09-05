package parser

import (
	"fmt"
	"github.com/rarnu/goscript/file"
	"github.com/rarnu/goscript/token"
	"sort"
)

const (
	errUnexpectedToken      = "Unexpected token %v"
	errUnexpectedEndOfInput = "Unexpected end of input"
	errUnexpectedEscape     = "Unexpected escape"
)

//    UnexpectedNumber:  'Unexpected number',
//    UnexpectedString:  'Unexpected string',
//    UnexpectedIdentifier:  'Unexpected identifier',
//    UnexpectedReserved:  'Unexpected reserved word',
//    NewlineAfterThrow:  'Illegal newline after throw',
//    InvalidRegExp: 'Invalid regular expression',
//    UnterminatedRegExp:  'Invalid regular expression: missing /',
//    InvalidLHSInAssignment:  'Invalid left-hand side in assignment',
//    InvalidLHSInForIn:  'Invalid left-hand side in for-in',
//    MultipleDefaultsInSwitch: 'More than one default clause in switch statement',
//    NoCatchOrFinally:  'Missing catch or finally after try',
//    UnknownLabel: 'Undefined label \'%0\'',
//    Redeclaration: '%0 \'%1\' has already been declared',
//    IllegalContinue: 'Illegal continue statement',
//    IllegalBreak: 'Illegal break statement',
//    IllegalReturn: 'Illegal return statement',
//    StrictModeWith:  'Strict mode code may not include a with statement',
//    StrictCatchVariable:  'Catch variable may not be eval or arguments in strict mode',
//    StrictVarName:  'Variable name may not be eval or arguments in strict mode',
//    StrictParamName:  'Parameter name eval or arguments is not allowed in strict mode',
//    StrictParamDupe: 'Strict mode function may not have duplicate parameter names',
//    StrictFunctionName:  'Function name may not be eval or arguments in strict mode',
//    StrictOctalLiteral:  'Octal literals are not allowed in strict mode.',
//    StrictDelete:  'Delete of an unqualified identifier in strict mode.',
//    StrictDuplicateProperty:  'Duplicate data property in object literal not allowed in strict mode',
//    AccessorDataProperty:  'Object literal may not have data and accessor property with the same name',
//    AccessorGetSet:  'Object literal may not have multiple get/set accessors with the same name',
//    StrictLHSAssignment:  'Assignment to eval or arguments is not allowed in strict mode',
//    StrictLHSPostfix:  'Postfix increment/decrement may not have eval or arguments operand in strict mode',
//    StrictLHSPrefix:  'Prefix increment/decrement may not have eval or arguments operand in strict mode',
//    StrictReservedWord:  'Use of future reserved word in strict mode'

// Error JavaScript 解析错误，它包含了发生错误的位置和具体的描述
type Error struct {
	Position file.Position
	Message  string
}

// ErrorList 错误列表
type ErrorList []*Error

func (e Error) Error() string {
	filename := e.Position.Filename
	if filename == "" {
		filename = "(anonymous)"
	}
	return fmt.Sprintf("%s: Line %d:%d %s",
		filename,
		e.Position.Line,
		e.Position.Column,
		e.Message,
	)
}

func (p *parser) error(place any, msg string, msgValues ...any) *Error {
	idx := file.Idx(0)
	switch pl := place.(type) {
	case int:
		idx = p.idxOf(pl)
	case file.Idx:
		if pl == 0 {
			idx = p.idxOf(p.chrOffset)
		} else {
			idx = pl
		}
	default:
		panic(fmt.Errorf("error(%T, ...)", pl))
	}
	position := p.position(idx)
	msg = fmt.Sprintf(msg, msgValues...)
	p.errors.Add(position, msg)
	return p.errors[len(p.errors)-1]
}

func (p *parser) errorUnexpected(idx file.Idx, chr rune) error {
	if chr == -1 {
		return p.error(idx, errUnexpectedEndOfInput)
	}
	return p.error(idx, errUnexpectedToken, token.ILLEGAL)
}

func (p *parser) errorUnexpectedToken(tkn token.Token) error {
	switch tkn {
	case token.EOF:
		return p.error(file.Idx(0), errUnexpectedEndOfInput)
	}
	value := tkn.String()
	switch tkn {
	case token.BOOLEAN, token.NULL:
		value = p.literal
	case token.IDENTIFIER:
		return p.error(p.idx, "Unexpected identifier")
	case token.KEYWORD:
		// 同样可能是未来的关键字
		return p.error(p.idx, "Unexpected reserved word")
	case token.ESCAPED_RESERVED_WORD:
		return p.error(p.idx, "Keyword must not contain escaped characters")
	case token.NUMBER:
		return p.error(p.idx, "Unexpected number")
	case token.STRING:
		return p.error(p.idx, "Unexpected string")
	}
	return p.error(p.idx, errUnexpectedToken, value)
}

func (l *ErrorList) Add(position file.Position, msg string) {
	*l = append(*l, &Error{position, msg})
}

func (l *ErrorList) Reset()        { *l = (*l)[0:0] }
func (l *ErrorList) Len() int      { return len(*l) }
func (l *ErrorList) Swap(i, j int) { (*l)[i], (*l)[j] = (*l)[j], (*l)[i] }
func (l *ErrorList) Less(i, j int) bool {
	x := &((*l)[i]).Position
	y := &((*l)[j]).Position
	if x.Filename < y.Filename {
		return true
	}
	if x.Filename == y.Filename {
		if x.Line < y.Line {
			return true
		}
		if x.Line == y.Line {
			return x.Column < y.Column
		}
	}
	return false
}

func (l *ErrorList) Sort() {
	sort.Sort(l)
}

func (l *ErrorList) Error() string {
	switch len(*l) {
	case 0:
		return "no errors"
	case 1:
		return (*l)[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", (*l)[0].Error(), len(*l)-1)
}

func (l *ErrorList) Err() error {
	if len(*l) == 0 {
		return nil
	}
	return l
}
