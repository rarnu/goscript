// Package unistring 是一个混合ASCII/UTF16字符串的实现
// 对于 ASCII 字符串，底层处理相当于普通的 Go 字符串
// 对于 unicode 字符串，底层处理为 []uint16 ，符合 UTF16 的编码规则，第0个元素为0xFEFF
package unistring

import (
	"reflect"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

const BOM = 0xFEFF

type String string

// Scan 检查一个字符串是否包含 unicode 字符，如果有，则转换为使用 FromUtf16 创建的字符串，否则返回 nil
func Scan(s string) []uint16 {
	hasUnicodeChar := false
	utf16Size := 0
	for ; utf16Size < len(s); utf16Size++ {
		if s[utf16Size] >= utf8.RuneSelf {
			hasUnicodeChar = true
			break
		}
	}
	if !hasUnicodeChar {
		return nil
	}
	for _, chr := range s[utf16Size:] {
		utf16Size++
		if chr > 0xFFFF {
			utf16Size++
		}
	}
	buf := make([]uint16, utf16Size+1)
	buf[0] = BOM
	c := 1
	for _, chr := range s {
		if chr <= 0xFFFF {
			buf[c] = uint16(chr)
		} else {
			first, second := utf16.EncodeRune(chr)
			buf[c] = uint16(first)
			c++
			buf[c] = uint16(second)
		}
		c++
	}
	return buf
}

func NewFromString(s string) String {
	if buf := Scan(s); buf != nil {
		return FromUtf16(buf)
	}
	return String(s)
}

func NewFromRunes(s []rune) String {
	ascii := true
	size := 0
	for _, c := range s {
		if c >= utf8.RuneSelf {
			ascii = false
			if c > 0xFFFF {
				size++
			}
		}
		size++
	}
	if ascii {
		return String(s)
	}
	b := make([]uint16, size+1)
	b[0] = BOM
	i := 1
	for _, c := range s {
		if c <= 0xFFFF {
			b[i] = uint16(c)
		} else {
			first, second := utf16.EncodeRune(c)
			b[i] = uint16(first)
			i++
			b[i] = uint16(second)
		}
		i++
	}
	return FromUtf16(b)
}

func FromUtf16(b []uint16) String {
	var str string
	hdr := (*reflect.StringHeader)(unsafe.Pointer(&str))
	hdr.Data = uintptr(unsafe.Pointer(&b[0]))
	hdr.Len = len(b) * 2
	return String(str)
}

func (s String) String() string {
	if b := s.AsUtf16(); b != nil {
		return string(utf16.Decode(b[1:]))
	}
	return string(s)
}

func (s String) AsUtf16() []uint16 {
	if len(s) < 4 || len(s)&1 != 0 {
		return nil
	}
	var a []uint16
	raw := string(s)
	sliceHeader := (*reflect.SliceHeader)(unsafe.Pointer(&a))
	sliceHeader.Data = (*reflect.StringHeader)(unsafe.Pointer(&raw)).Data
	l := len(raw) / 2
	sliceHeader.Len = l
	sliceHeader.Cap = l
	if a[0] == BOM {
		return a
	}
	return nil
}
