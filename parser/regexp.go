package parser

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	WhitespaceChars = " \f\n\r\t\v\u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a\u2028\u2029\u202f\u205f\u3000\ufeff"
	Re2Dot          = "[^\r\n\u2028\u2029]"
)

type regexpParseError struct {
	offset int
	err    string
}

type RegexpErrorIncompatible struct {
	regexpParseError
}
type RegexpSyntaxError struct {
	regexpParseError
}

func (s regexpParseError) Error() string {
	return s.err
}

type regExpParser struct {
	str        string
	length     int
	chr        rune // 当前字符
	chrOffset  int  // 当前字符的偏移量(offset)
	offset     int  // 紧随当前字符的偏移量，大于等于 1
	err        error
	goRegexp   strings.Builder
	passOffset int
}

// TransformRegExp 将一个 JavaScript 的正则表达式转换为 Go 的正则
// re2 (Go) 不能进行回溯，因此遇到回看，如 (?=) (?!) 或 (\1, \2, ...) 将报错
// re2 (Go) 对 \s 有不同的定义： [\t\n\f\r ]
// 另外，JavaScript 还包含了额外的 \v, Unicode 的分割符和空格等
func TransformRegExp(pattern string) (transformed string, err error) {
	if pattern == "" {
		return "", nil
	}
	par := regExpParser{str: pattern, length: len(pattern)}
	err = par.parse()
	if err != nil {
		return "", err
	}
	return par.ResultString(), nil
}

func (r *regExpParser) ResultString() string {
	if r.passOffset != -1 {
		return r.str[:r.passOffset]
	}
	return r.goRegexp.String()
}

func (r *regExpParser) parse() (err error) {
	r.read() // 读取第一个字符
	r.scan()
	return r.err
}

func (r *regExpParser) read() {
	if r.offset < r.length {
		r.chrOffset = r.offset
		chr, width := rune(r.str[r.offset]), 1
		if chr >= utf8.RuneSelf {
			// 不是 ASCII 字符
			chr, width = utf8.DecodeRuneInString(r.str[r.offset:])
			if chr == utf8.RuneError && width == 1 {
				r.error(true, "Invalid UTF-8 character")
				return
			}
		}
		r.offset += width
		r.chr = chr
	} else {
		r.chrOffset = r.length
		r.chr = -1 // 读到末尾了
	}
}

func (r *regExpParser) stopPassing() {
	r.goRegexp.Grow(3 * len(r.str) / 2)
	r.goRegexp.WriteString(r.str[:r.passOffset])
	r.passOffset = -1
}

func (r *regExpParser) write(p []byte) {
	if r.passOffset != -1 {
		r.stopPassing()
	}
	r.goRegexp.Write(p)
}

func (r *regExpParser) writeByte(b byte) {
	if r.passOffset != -1 {
		r.stopPassing()
	}
	r.goRegexp.WriteByte(b)
}

func (r *regExpParser) writeString(s string) {
	if r.passOffset != -1 {
		r.stopPassing()
	}
	r.goRegexp.WriteString(s)
}

func (r *regExpParser) scan() {
	for r.chr != -1 {
		switch r.chr {
		case '\\':
			r.read()
			r.scanEscape(false)
		case '(':
			r.pass()
			r.scanGroup()
		case '[':
			r.scanBracket()
		case ')':
			r.error(true, "Unmatched ')'")
			return
		case '.':
			r.writeString(Re2Dot)
			r.read()
		default:
			r.pass()
		}
	}
}

func (r *regExpParser) scanGroup() {
	str := r.str[r.chrOffset:]
	if len(str) > 1 {
		// 此处可能出现 (?= 或 (?!
		if str[0] == '?' {
			ch := str[1]
			switch {
			case ch == '=' || ch == '!':
				r.error(false, "re2: Invalid (%s) <lookahead>", r.str[r.chrOffset:r.chrOffset+2])
				return
			case ch == '<':
				r.error(false, "re2: Invalid (%s) <lookbehind>", r.str[r.chrOffset:r.chrOffset+2])
				return
			case ch != ':':
				r.error(true, "Invalid group")
				return
			}
		}
	}
	for r.chr != -1 && r.chr != ')' {
		switch r.chr {
		case '\\':
			r.read()
			r.scanEscape(false)
		case '(':
			r.pass()
			r.scanGroup()
		case '[':
			r.scanBracket()
		case '.':
			r.writeString(Re2Dot)
			r.read()
		default:
			r.pass()
			continue
		}
	}
	if r.chr != ')' {
		r.error(true, "Unterminated group")
		return
	}
	r.pass()
}

func (r *regExpParser) scanBracket() {
	str := r.str[r.chrOffset:]
	if strings.HasPrefix(str, "[]") {
		r.writeString("[^\u0000-\U0001FFFF]")
		r.offset += 1
		r.read()
		return
	}
	if strings.HasPrefix(str, "[^]") {
		r.writeString("[\u0000-\U0001FFFF]")
		r.offset += 2
		r.read()
		return
	}
	r.pass()
	for r.chr != -1 {
		if r.chr == ']' {
			break
		} else if r.chr == '\\' {
			r.read()
			r.scanEscape(true)
			continue
		}
		r.pass()
	}
	if r.chr != ']' {
		r.error(true, "Unterminated character class")
		return
	}
	r.pass()
}

func (r *regExpParser) scanEscape(inClass bool) {
	offset := r.chrOffset
	var length, base uint32
	switch r.chr {
	case '0', '1', '2', '3', '4', '5', '6', '7':
		var value int64
		size := 0
		for {
			digit := int64(digitValue(r.chr))
			if digit >= 8 {
				// 非法的位数
				break
			}
			value = value*8 + digit
			r.read()
			size += 1
		}
		if size == 1 {
			if value != 0 {
				r.error(false, "re2: Invalid \\%d <backreference>", value)
				return
			}
			r.passString(offset-1, r.chrOffset)
			return
		}
		tmp := []byte{'\\', 'x', '0', 0}
		if value >= 16 {
			tmp = tmp[0:2]
		} else {
			tmp = tmp[0:3]
		}
		tmp = strconv.AppendInt(tmp, value, 16)
		r.write(tmp)
		return

	case '8', '9':
		r.read()
		r.error(false, "re2: Invalid \\%s <backreference>", r.str[offset:r.chrOffset])
		return

	case 'x':
		r.read()
		length, base = 2, 16

	case 'u':
		r.read()
		if r.chr == '{' {
			r.read()
			length, base = 0, 16
		} else {
			length, base = 4, 16
		}

	case 'b':
		if inClass {
			r.write([]byte{'\\', 'x', '0', '8'})
			r.read()
			return
		}
		fallthrough

	case 'B':
		fallthrough

	case 'd', 'D', 'w', 'W':
		// 这里会有问题，因为ECMAScript在 \s,\S 内包含了 \v，但是 re2 不包含
		fallthrough

	case '\\':
		fallthrough

	case 'f', 'n', 'r', 't', 'v':
		r.passString(offset-1, r.offset)
		r.read()
		return

	case 'c':
		r.read()
		var value int64
		if 'a' <= r.chr && r.chr <= 'z' {
			value = int64(r.chr - 'a' + 1)
		} else if 'A' <= r.chr && r.chr <= 'Z' {
			value = int64(r.chr - 'A' + 1)
		} else {
			r.writeByte('c')
			return
		}
		tmp := []byte{'\\', 'x', '0', 0}
		if value >= 16 {
			tmp = tmp[0:2]
		} else {
			tmp = tmp[0:3]
		}
		tmp = strconv.AppendInt(tmp, value, 16)
		r.write(tmp)
		r.read()
		return
	case 's':
		if inClass {
			r.writeString(WhitespaceChars)
		} else {
			r.writeString("[" + WhitespaceChars + "]")
		}
		r.read()
		return
	case 'S':
		if inClass {
			r.error(false, "S in class")
			return
		} else {
			r.writeString("[^" + WhitespaceChars + "]")
		}
		r.read()
		return
	default:
		if r.chr == '$' || r.chr < utf8.RuneSelf && !isIdentifierPart(r.chr) {
			r.passString(offset-1, r.offset)
			r.read()
			return
		}
		r.pass()
		return
	}

	valueOffset := r.chrOffset

	if length > 0 {
		for length := length; length > 0; length-- {
			digit := uint32(digitValue(r.chr))
			if digit >= base {
				// 非法的位数
				goto skip
			}
			r.read()
		}
	} else {
		for r.chr != '}' && r.chr != -1 {
			digit := uint32(digitValue(r.chr))
			if digit >= base {
				// 非法的位数
				goto skip
			}
			r.read()
		}
	}
	if length == 4 || length == 0 {
		r.write([]byte{
			'\\',
			'x',
			'{',
		})
		r.passString(valueOffset, r.chrOffset)
		if length != 0 {
			r.writeByte('}')
		}
	} else if length == 2 {
		r.passString(offset-1, valueOffset+2)
	} else {
		r.error(true, "re2: Illegal branch in scanEscape")
		return
	}
	return
skip:
	r.passString(offset, r.chrOffset)
}

func (r *regExpParser) pass() {
	if r.passOffset == r.chrOffset {
		r.passOffset = r.offset
	} else {
		if r.passOffset != -1 {
			r.stopPassing()
		}
		if r.chr != -1 {
			r.goRegexp.WriteRune(r.chr)
		}
	}
	r.read()
}

func (r *regExpParser) passString(start, end int) {
	if r.passOffset == start {
		r.passOffset = end
		return
	}
	if r.passOffset != -1 {
		r.stopPassing()
	}
	r.goRegexp.WriteString(r.str[start:end])
}

func (r *regExpParser) error(fatal bool, msg string, msgValues ...any) {
	if r.err != nil {
		return
	}
	e := regexpParseError{
		offset: r.offset,
		err:    fmt.Sprintf(msg, msgValues...),
	}
	if fatal {
		r.err = RegexpSyntaxError{e}
	} else {
		r.err = RegexpErrorIncompatible{e}
	}
	r.offset = r.length
	r.chr = -1
}
