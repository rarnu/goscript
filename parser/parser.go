package parser

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/rarnu/goscript/ast"
	"github.com/rarnu/goscript/file"
	"github.com/rarnu/goscript/token"
	"github.com/rarnu/goscript/unistring"
	"io"
	"os"
)

// Mode 保存一组标志，该标志用于控制 parser 的可选功能
type Mode uint

const IgnoreRegExpErrors Mode = 1 << iota // 忽略正则兼容性错误 (允许回溯)

type options struct {
	disableSourceMaps bool
	sourceMapLoader   func(path string) ([]byte, error)
}

// Option 表示在 parser 内使用的选项之一，目前支持的是 WithDisableSourceMaps 和 WithSourceMapLoader
type Option func(*options)

// WithDisableSourceMaps 禁用源码 map 的选项，在不使用源码 map 时，将会节省解析 map 的时间
func WithDisableSourceMaps(opts *options) {
	opts.disableSourceMaps = true
}

// WithSourceMapLoader 启用源码 map 的选项，可以设置一个自定义的源码 map 加载器
// 该加载器将被赋予一个路径或一个 sourceMappingURL 中的 URL，sourceMappingURL 可以是绝对路径或相对路径
// 任何由加载器返回的错误都会导致解析失败
func WithSourceMapLoader(loader func(path string) ([]byte, error)) Option {
	return func(opts *options) {
		opts.sourceMapLoader = loader
	}
}

type parser struct {
	str               string
	length            int
	base              int
	chr               rune        // 当前的字符
	chrOffset         int         // 当前字符的偏移量(offset)
	offset            int         // 紧随当前字符的偏移量(offset), 必定大于或等于 1
	idx               file.Idx    // 令牌(token)的索引
	token             token.Token // 令牌(token)
	literal           string      // 令牌(token)的字面量(literal) (哪果有的话)
	parsedLiteral     unistring.String
	scope             *scope
	insertSemicolon   bool // 如果发现有换行，则插入一个隐藏的分号
	implicitSemicolon bool // 隐藏的分号是否已存在
	errors            ErrorList
	recover           struct {
		// 在试图寻找下一个语句时，出现未预期的异常 (如不完整的语句等)
		idx   file.Idx
		count int
	}
	mode Mode
	opts options
	file *file.File
}

func _newParser(filename, src string, base int, opts ...Option) *parser {
	p := &parser{
		chr:    ' ', // 设置跳过起始的空白字符，然后进行扫描
		str:    src,
		length: len(src),
		base:   base,
		file:   file.NewFile(filename, src, base),
	}
	for _, opt := range opts {
		opt(&p.opts)
	}
	return p
}

func newParser(filename, src string) *parser {
	return _newParser(filename, src, 1)
}

func ReadSource(filename string, src any) ([]byte, error) {
	if src != nil {
		switch src := src.(type) {
		case string:
			return []byte(src), nil
		case []byte:
			return src, nil
		case *bytes.Buffer:
			if src != nil {
				return src.Bytes(), nil
			}
		case io.Reader:
			var bfr bytes.Buffer
			if _, err := io.Copy(&bfr, src); err != nil {
				return nil, err
			}
			return bfr.Bytes(), nil
		}
		return nil, errors.New("invalid source")
	}
	return os.ReadFile(filename)
}

// ParseFile 解析单个 JavaScript 源文件的代码并返回相应的 ast.Program 节点
// 若文件集合为空，则解析时不包含文件集合(没有上下文)，若文件集合非空，则先将要解析的文件加入其中
// 文件名参数可选，用于在出现错误时，标记错误发生的文件名
// src 可以是 string/[]byte/bytes.Buffer/io.Reader, 但是不论如何，它的内容编码必须是UTF8
// 解析 JavaScript 将最终产生一个 *ast.Program 和一个 ErrorList
func ParseFile(fileSet *file.FileSet, filename string, src any, mode Mode, options ...Option) (*ast.Program, error) {
	b, err := ReadSource(filename, src)
	if err != nil {
		return nil, err
	}
	str := string(b)
	base := 1
	if fileSet != nil {
		base = fileSet.AddFile(filename, str)
	}
	p := _newParser(filename, str, base, options...)
	p.mode = mode
	return p.parse()

}

// ParseFunction 将一个给定的参数列表和主体解析为一个函数，并返回相应的 ast.FunctionLiteral 节点
// 若有参数，则参数列表必须是一个用逗号分隔的标识符列表
func ParseFunction(parameterList, body string, options ...Option) (*ast.FunctionLiteral, error) {
	src := fmt.Sprintf("(function(%s) {\n%s\n})", parameterList, body)
	p := _newParser("", src, 1, options...)
	program, err := p.parse()
	if err != nil {
		return nil, err
	}
	return program.Body[0].(*ast.ExpressionStatement).Expression.(*ast.FunctionLiteral), nil
}

func (p *parser) slice(idx0, idx1 file.Idx) string {
	from := int(idx0) - p.base
	to := int(idx1) - p.base
	if from >= 0 && to <= len(p.str) {
		return p.str[from:to]
	}
	return ""
}

func (p *parser) parse() (*ast.Program, error) {
	p.next()
	program := p.parseProgram()
	if false {
		p.errors.Sort()
	}
	return program, p.errors.Err()
}

func (p *parser) next() {
	p.token, p.literal, p.parsedLiteral, p.idx = p.scan()
}

func (p *parser) optionalSemicolon() {
	if p.token == token.SEMICOLON {
		p.next()
		return
	}
	if p.implicitSemicolon {
		p.implicitSemicolon = false
		return
	}
	if p.token != token.EOF && p.token != token.RIGHT_BRACE {
		p.expect(token.SEMICOLON)
	}
}

func (p *parser) semicolon() {
	if p.token != token.RIGHT_PARENTHESIS && p.token != token.RIGHT_BRACE {
		if p.implicitSemicolon {
			p.implicitSemicolon = false
			return
		}
		p.expect(token.SEMICOLON)
	}
}

func (p *parser) idxOf(offset int) file.Idx {
	return file.Idx(p.base + offset)
}

func (p *parser) expect(value token.Token) file.Idx {
	idx := p.idx
	if p.token != value {
		_ = p.errorUnexpectedToken(p.token)
	}
	p.next()
	return idx
}

func (p *parser) position(idx file.Idx) file.Position {
	return p.file.Position(int(idx) - p.base)
}
