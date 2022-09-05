package require

import (
	"errors"
	js "github.com/rarnu/goscript"
	"github.com/rarnu/goscript/parser"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"text/template"
)

type ModuleLoader func(*js.Runtime, *js.Object)

// SourceLoader 是一个函数，它在给定的路径上返回一个文件数据
// 如果文件不存在或者是一个目录，返回 ModuleFileDoesNotExistError
// 这个错误将被解析器忽略，搜索将继续进行。任何其他错误都会被传递
type SourceLoader func(path string) ([]byte, error)

var (
	InvalidModuleError          = errors.New("invalid module")
	IllegalModuleNameError      = errors.New("illegal module name")
	ModuleFileDoesNotExistError = errors.New("module file does not exist")
)

var native map[string]ModuleLoader

// Registry 包含一个可供多个运行时使用的已编译的模块缓存
type Registry struct {
	sync.Mutex
	native        map[string]ModuleLoader
	compiled      map[string]*js.Program
	srcLoader     SourceLoader
	globalFolders []string
}

type RequireModule struct {
	r           *Registry
	runtime     *js.Runtime
	modules     map[string]*js.Object
	nodeModules map[string]*js.Object
}

func NewRegistry(opts ...Option) *Registry {
	r := &Registry{}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func NewRegistryWithLoader(srcLoader SourceLoader) *Registry {
	return NewRegistry(WithLoader(srcLoader))
}

type Option func(*Registry)

// WithLoader 设置了一个函数，这个函数将被 require() 函数调用，以便在给定的路径上获得一个模块的源代码
// 同样的，函数将被用来获取外部的源码 map
// 注意，这只影响由 require()函 数加载的模块
// 如果你需要把它作为源码 map 加载器，用于以不同方式解析的代码（如 runtime.RunString() 或 eval() ），请使用 (*Runtime).SetParserOptions()
func WithLoader(srcLoader SourceLoader) Option {
	return func(r *Registry) {
		r.srcLoader = srcLoader
	}
}

// WithGlobalFolders 将给定的路径添加到 Registry 的全局文件夹列表中，以便在其他地方找不到请求的模块时进行搜索。 默认情况下，一个 Registry 的全局文件夹列表是空的
// 参考的 Nodejs 的实现，默认的全局文件夹列表是 $NODE_PATH、$HOME/.node_modules、$HOME/.node_libraries 和 $PREFIX/lib/node
// 参考: https://nodejs.org/api/modules.html#modules_loading_from_the_global_folders
func WithGlobalFolders(globalFolders ...string) Option {
	return func(r *Registry) {
		r.globalFolders = globalFolders
	}
}

// Enable 将 require() 函数添加到指定的 Runtime
func (r *Registry) Enable(runtime *js.Runtime) *RequireModule {
	rrt := &RequireModule{
		r:           r,
		runtime:     runtime,
		modules:     make(map[string]*js.Object),
		nodeModules: make(map[string]*js.Object),
	}
	_ = runtime.Set("require", rrt.require)
	return rrt
}

func (r *Registry) RegisterNativeModule(name string, loader ModuleLoader) {
	r.Lock()
	defer r.Unlock()
	if r.native == nil {
		r.native = make(map[string]ModuleLoader)
	}
	name = filepathClean(name)
	r.native[name] = loader
}

// DefaultSourceLoader 如果没有设置 SourceLoader，就使用 DefaultSourceLoader（参考 WithLoader() ），它只会从宿主文件系统中加载文件
func DefaultSourceLoader(filename string) ([]byte, error) {
	fp := filepath.FromSlash(filename)
	data, err := ioutil.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) || errors.Is(err, syscall.EISDIR) {
			err = ModuleFileDoesNotExistError
		} else if runtime.GOOS == "windows" {
			// 针对 windows 的一个变通方法
			fi, err1 := os.Stat(fp)
			if err1 == nil && fi.IsDir() {
				err = ModuleFileDoesNotExistError
			}
		}
	}
	return data, err
}

func (r *Registry) getSource(p string) ([]byte, error) {
	srcLoader := r.srcLoader
	if srcLoader == nil {
		srcLoader = DefaultSourceLoader
	}
	return srcLoader(p)
}

func (r *Registry) getCompiledSource(p string) (*js.Program, error) {
	r.Lock()
	defer r.Unlock()

	prg := r.compiled[p]
	if prg == nil {
		buf, err := r.getSource(p)
		if err != nil {
			return nil, err
		}
		s := string(buf)

		if path.Ext(p) == ".json" {
			s = "module.exports = JSON.parse('" + template.JSEscapeString(s) + "')"
		}

		source := "(function(exports, require, module) {" + s + "\n})"
		parsed, err := js.Parse(p, source, parser.WithSourceMapLoader(r.srcLoader))
		if err != nil {
			return nil, err
		}
		prg, err = js.CompileAST(parsed, false)
		if err == nil {
			if r.compiled == nil {
				r.compiled = make(map[string]*js.Program)
			}
			r.compiled[p] = prg
		}
		return prg, err
	}
	return prg, nil
}

func (r *RequireModule) require(call js.FunctionCall) js.Value {
	ret, err := r.Require(call.Argument(0).String())
	if err != nil {
		if _, ok := err.(*js.Exception); !ok {
			panic(r.runtime.NewGoError(err))
		}
		panic(err)
	}
	return ret
}

func filepathClean(p string) string {
	return path.Clean(p)
}

// Require 可以用于从 Go 源代码中导入模块（类似于 JS 的 require() 函数）
func (r *RequireModule) Require(p string) (ret js.Value, err error) {
	module, err := r.resolve(p)
	if err != nil {
		return
	}
	ret = module.Get("exports")
	return
}

func Require(runtime *js.Runtime, name string) js.Value {
	if r, ok := js.AssertFunction(runtime.Get("require")); ok {
		mod, err := r(js.Undefined(), runtime.ToValue(name))
		if err != nil {
			panic(err)
		}
		return mod
	}
	panic(runtime.NewTypeError("Please enable require for this runtime using new(require.Registry).Enable(runtime)"))
}

func RegisterNativeModule(name string, loader ModuleLoader) {
	if native == nil {
		native = make(map[string]ModuleLoader)
	}
	name = filepathClean(name)
	native[name] = loader
}
