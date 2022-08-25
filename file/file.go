// Package file 供 AST 使用的文件操作抽象
package file

import (
	"fmt"
	"github.com/go-sourcemap/sourcemap"
	"net/url"
	"path"
	"sort"
	"sync"
)

// Idx 是在一个在文件集合内的。指出源码位置的一个紧凑型编码
// 它可以被转换为 Position，以获得一个更为全面的数据表达
type Idx int

// Position 用于描述一个任意的源码位置，包括文件名，行和列
type Position struct {
	Filename string // 产生了错误的文件名（有错误发生的情况才填充）
	Line     int    // 行号，以1为起始
	Column   int    // 列号，以1为起始
}

// FileSet 代表了一组源文件
type FileSet struct {
	files []*File
	last  *File
}

// File 代表了一个具体的源文件
type File struct {
	mu                sync.Mutex
	name              string
	src               string
	base              int // 1 或更大的值
	sourceMap         *sourcemap.Consumer
	lineOffsets       []int
	lastScannedOffset int
}

// isValid 判断 Position 是否合法
func (p *Position) isValid() bool {
	// 当 Line > 0 时，宣告 Position 合法
	return p.Line > 0
}

func (p *Position) String() string {
	str := p.Filename
	if p.isValid() {
		if str != "" {
			str += ":"
		}
		str += fmt.Sprintf("%d:%d", p.Line, p.Column)
	}
	if str == "" {
		str = "-"
	}
	return str
}

// AddFile 将指定文件名和源文件路径的一个文件，添加到文件集合中
func (fs *FileSet) AddFile(filename, src string) int {
	base := fs.nextBase()
	file := &File{
		name: filename,
		src:  src,
		base: base,
	}
	fs.files = append(fs.files, file)
	fs.last = file
	return base
}

func (fs *FileSet) nextBase() int {
	if fs.last == nil {
		return 1
	}
	return fs.last.base + len(fs.last.src) + 1
}

// File 根据下标获取具体的文件
func (fs *FileSet) File(idx Idx) *File {
	for _, file := range fs.files {
		if idx <= Idx(file.base+len(file.src)) {
			return file
		}
	}
	return nil
}

// Position 将在文件集合中的 Idx 转换为 Position
func (fs *FileSet) Position(idx Idx) Position {
	for _, file := range fs.files {
		if idx <= Idx(file.base+len(file.src)) {
			return file.Position(int(idx) - file.base)
		}
	}
	return Position{}
}

// NewFile 新建一个文件
func NewFile(filename, src string, base int) *File {
	return &File{
		name: filename,
		src:  src,
		base: base,
	}
}

func (f *File) Name() string {
	return f.name
}

func (f *File) Source() string {
	return f.src
}

func (f *File) Base() int {
	return f.base
}

func (f *File) SetSourceMap(m *sourcemap.Consumer) {
	f.sourceMap = m
}

// Position 根据文件指针的偏移量，计算出该指针所在的具体位置
func (f *File) Position(offset int) Position {
	var line int
	var lineOffsets []int
	f.mu.Lock()
	if offset > f.lastScannedOffset {
		line = f.scanTo(offset)
		lineOffsets = f.lineOffsets
		f.mu.Unlock()
	} else {
		lineOffsets = f.lineOffsets
		f.mu.Unlock()
		line = sort.Search(len(lineOffsets), func(x int) bool { return lineOffsets[x] > offset }) - 1
	}
	var lineStart int
	if line >= 0 {
		lineStart = lineOffsets[line]
	}
	row := line + 2
	col := offset - lineStart + 1
	if f.sourceMap != nil {
		if source, _, row, col, ok := f.sourceMap.Source(row, col); ok {
			return Position{
				Filename: ResolveSourcemapURL(f.Name(), source).String(),
				Line:     row,
				Column:   col,
			}
		}
	}
	return Position{
		Filename: f.name,
		Line:     row,
		Column:   col,
	}
}

func ResolveSourcemapURL(basename, source string) *url.URL {
	smURL, err := url.Parse(source)
	if err == nil && !smURL.IsAbs() {
		baseURL, err1 := url.Parse(basename)
		if err1 == nil && path.IsAbs(baseURL.Path) {
			smURL = baseURL.ResolveReference(smURL)
		} else {
			smURL, _ = url.Parse(path.Join(path.Dir(basename), smURL.Path))
		}
	}
	return smURL
}

func findNextLineStart(s string) int {
	for pos, ch := range s {
		switch ch {
		case '\r':
			if pos < len(s)-1 && s[pos+1] == '\n' {
				return pos + 2
			}
			return pos + 1
		case '\n':
			return pos + 1
		case '\u2028', '\u2029':
			return pos + 3
		}
	}
	return -1
}

func (f *File) scanTo(offset int) int {
	o := f.lastScannedOffset
	for o < offset {
		p := findNextLineStart(f.src[o:])
		if p == -1 {
			f.lastScannedOffset = len(f.src)
			return len(f.lineOffsets) - 1
		}
		o = o + p
		f.lineOffsets = append(f.lineOffsets, o)
	}
	f.lastScannedOffset = o
	if o == offset {
		return len(f.lineOffsets) - 1
	}
	return len(f.lineOffsets) - 2
}
