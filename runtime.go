package goscript

import (
	"bytes"
	"errors"
	"fmt"
	jast "github.com/rarnu/goscript/ast"
	"github.com/rarnu/goscript/file"
	"github.com/rarnu/goscript/parser"
	"github.com/rarnu/goscript/unistring"
	"go/ast"
	"golang.org/x/text/collate"
	"hash/maphash"
	"math"
	"math/bits"
	"math/rand"
	"reflect"
	"runtime"
	"strconv"
	"time"
)

const (
	sqrt1_2          float64 = math.Sqrt2 / 2
	deoptimiseRegexp         = false
)

var (
	typeCallable = reflect.TypeOf(Callable(nil))
	typeValue    = reflect.TypeOf((*Value)(nil)).Elem()
	typeObject   = reflect.TypeOf((*Object)(nil))
	typeTime     = reflect.TypeOf(time.Time{})
	typeBytes    = reflect.TypeOf(([]byte)(nil))
)

type iterationKind int

const (
	iterationKindKey iterationKind = iota
	iterationKindValue
	iterationKindKeyValue
)

type global struct {
	stash    stash
	varNames map[unistring.String]struct{}

	Object     *Object
	Array      *Object
	Function   *Object
	String     *Object
	Number     *Object
	Boolean    *Object
	RegExp     *Object
	Date       *Object
	Mysql      *Object
	Redis      *Object
	Etcd       *Object
	Kubernetes *Object
	Symbol     *Object
	Proxy      *Object
	Promise    *Object

	ArrayBuffer       *Object
	DataView          *Object
	TypedArray        *Object
	Uint8Array        *Object
	Uint8ClampedArray *Object
	Int8Array         *Object
	Uint16Array       *Object
	Int16Array        *Object
	Uint32Array       *Object
	Int32Array        *Object
	Float32Array      *Object
	Float64Array      *Object

	WeakSet *Object
	WeakMap *Object
	Map     *Object
	Set     *Object

	Error          *Object
	AggregateError *Object
	TypeError      *Object
	ReferenceError *Object
	SyntaxError    *Object
	RangeError     *Object
	EvalError      *Object
	URIError       *Object

	GoError *Object

	ObjectPrototype     *Object
	ArrayPrototype      *Object
	NumberPrototype     *Object
	StringPrototype     *Object
	BooleanPrototype    *Object
	FunctionPrototype   *Object
	RegExpPrototype     *Object
	DatePrototype       *Object
	SymbolPrototype     *Object
	MysqlPrototype      *Object
	RedisPrototype      *Object
	EtcdPrototype       *Object
	KubernetesPrototype *Object

	ArrayBufferPrototype *Object
	DataViewPrototype    *Object
	TypedArrayPrototype  *Object
	WeakSetPrototype     *Object
	WeakMapPrototype     *Object
	MapPrototype         *Object
	SetPrototype         *Object
	PromisePrototype     *Object

	IteratorPrototype             *Object
	ArrayIteratorPrototype        *Object
	MapIteratorPrototype          *Object
	SetIteratorPrototype          *Object
	StringIteratorPrototype       *Object
	RegExpStringIteratorPrototype *Object

	ErrorPrototype          *Object
	AggregateErrorPrototype *Object
	TypeErrorPrototype      *Object
	SyntaxErrorPrototype    *Object
	RangeErrorPrototype     *Object
	ReferenceErrorPrototype *Object
	EvalErrorPrototype      *Object
	URIErrorPrototype       *Object

	GoErrorPrototype *Object

	Eval *Object

	thrower         *Object
	throwerProperty Value

	stdRegexpProto *guardedObject

	weakSetAdder  *Object
	weakMapAdder  *Object
	mapAdder      *Object
	setAdder      *Object
	arrayValues   *Object
	arrayToString *Object
}

type Flag int

const (
	FLAG_NOT_SET Flag = iota
	FLAG_FALSE
	FLAG_TRUE
)

func (f Flag) Bool() bool {
	return f == FLAG_TRUE
}

func ToFlag(b bool) Flag {
	if b {
		return FLAG_TRUE
	}
	return FLAG_FALSE
}

type RandSource func() float64

type Now func() time.Time

type Runtime struct {
	global                  global
	globalObject            *Object
	stringSingleton         *stringObject
	rand                    RandSource
	now                     Now
	_collator               *collate.Collator
	parserOptions           []parser.Option
	symbolRegistry          map[unistring.String]*Symbol
	typeInfoCache           map[reflect.Type]*reflectTypeInfo
	fieldNameMapper         FieldNameMapper
	vm                      *vm
	hash                    *maphash.Hash
	idSeq                   uint64
	jobQueue                []func()
	promiseRejectionTracker PromiseRejectionTracker
}

type StackFrame struct {
	prg      *Program
	funcName unistring.String
	pc       int
}

func (f *StackFrame) SrcName() string {
	if f.prg == nil {
		return "<native>"
	}
	return f.prg.src.Name()
}

func (f *StackFrame) FuncName() string {
	if f.funcName == "" && f.prg == nil {
		return "<native>"
	}
	if f.funcName == "" {
		return "<anonymous>"
	}
	return f.funcName.String()
}

func (f *StackFrame) Position() file.Position {
	if f.prg == nil || f.prg.src == nil {
		return file.Position{}
	}
	return f.prg.src.Position(f.prg.sourceOffset(f.pc))
}

func (f *StackFrame) WriteToValueBuilder(b *valueStringBuilder) {
	if f.prg != nil {
		if n := f.prg.funcName; n != "" {
			b.WriteString(stringValueFromRaw(n))
			b.WriteASCII(" (")
		}
		p := f.Position()
		if p.Filename != "" {
			b.WriteASCII(p.Filename)
		} else {
			b.WriteASCII("<eval>")
		}
		b.WriteRune(':')
		b.WriteASCII(strconv.Itoa(p.Line))
		b.WriteRune(':')
		b.WriteASCII(strconv.Itoa(p.Column))
		b.WriteRune('(')
		b.WriteASCII(strconv.Itoa(f.pc))
		b.WriteRune(')')
		if f.prg.funcName != "" {
			b.WriteRune(')')
		}
	} else {
		if f.funcName != "" {
			b.WriteString(stringValueFromRaw(f.funcName))
			b.WriteASCII(" (")
		}
		b.WriteASCII("native")
		if f.funcName != "" {
			b.WriteRune(')')
		}
	}
}

func (f *StackFrame) Write(b *bytes.Buffer) {
	if f.prg != nil {
		if n := f.prg.funcName; n != "" {
			b.WriteString(n.String())
			b.WriteString(" (")
		}
		p := f.Position()
		if p.Filename != "" {
			b.WriteString(p.Filename)
		} else {
			b.WriteString("<eval>")
		}
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(p.Line))
		b.WriteByte(':')
		b.WriteString(strconv.Itoa(p.Column))
		b.WriteByte('(')
		b.WriteString(strconv.Itoa(f.pc))
		b.WriteByte(')')
		if f.prg.funcName != "" {
			b.WriteByte(')')
		}
	} else {
		if f.funcName != "" {
			b.WriteString(f.funcName.String())
			b.WriteString(" (")
		}
		b.WriteString("native")
		if f.funcName != "" {
			b.WriteByte(')')
		}
	}
}

type Exception struct {
	val   Value
	stack []StackFrame
}

type uncatchableException struct {
	err error
}

func (ue *uncatchableException) Unwrap() error {
	return ue.err
}

type InterruptedError struct {
	Exception
	iface any
}

func (e *InterruptedError) Unwrap() error {
	if err, ok := e.iface.(error); ok {
		return err
	}
	return nil
}

type StackOverflowError struct {
	Exception
}

func (e *InterruptedError) Value() any {
	return e.iface
}

func (e *InterruptedError) String() string {
	if e == nil {
		return "<nil>"
	}
	var b bytes.Buffer
	if e.iface != nil {
		b.WriteString(fmt.Sprint(e.iface))
		b.WriteByte('\n')
	}
	e.writeFullStack(&b)
	return b.String()
}

func (e *InterruptedError) Error() string {
	if e == nil || e.iface == nil {
		return "<nil>"
	}
	var b bytes.Buffer
	b.WriteString(fmt.Sprint(e.iface))
	e.writeShortStack(&b)
	return b.String()
}

func (e *Exception) writeFullStack(b *bytes.Buffer) {
	for _, frame := range e.stack {
		b.WriteString("\tat ")
		frame.Write(b)
		b.WriteByte('\n')
	}
}

func (e *Exception) writeShortStack(b *bytes.Buffer) {
	if len(e.stack) > 0 && (e.stack[0].prg != nil || e.stack[0].funcName != "") {
		b.WriteString(" at ")
		e.stack[0].Write(b)
	}
}

func (e *Exception) String() string {
	if e == nil {
		return "<nil>"
	}
	var b bytes.Buffer
	if e.val != nil {
		b.WriteString(e.val.String())
		b.WriteByte('\n')
	}
	e.writeFullStack(&b)
	return b.String()
}

func (e *Exception) Error() string {
	if e == nil || e.val == nil {
		return "<nil>"
	}
	var b bytes.Buffer
	b.WriteString(e.val.String())
	e.writeShortStack(&b)
	return b.String()
}

func (e *Exception) Value() Value {
	return e.val
}

func (r *Runtime) addToGlobal(name string, value Value) {
	r.globalObject.self._putProp(unistring.String(name), value, true, false, true)
}

func (r *Runtime) createIterProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putSym(SymIterator, valueProp(r.newNativeFunc(r.returnThis, nil, "[Symbol.iterator]", nil, 0), true, false, true))
	return o
}

func (r *Runtime) init() {
	r.rand = rand.Float64
	r.now = time.Now
	r.global.ObjectPrototype = r.newBaseObject(nil, classObject).val
	r.globalObject = r.NewObject()

	r.vm = &vm{
		r: r,
	}
	r.vm.init()

	funcProto := r.newNativeFunc(func(FunctionCall) Value {
		return _undefined
	}, nil, " ", nil, 0)
	r.global.FunctionPrototype = funcProto
	funcProtoObj := funcProto.self.(*nativeFuncObject)

	r.global.IteratorPrototype = r.newLazyObject(r.createIterProto)

	r.initObject()
	r.initFunction()
	r.initArray()
	r.initString()
	r.initGlobalObject()
	r.initNumber()
	r.initRegExp()
	r.initDate()
	r.initBoolean()
	r.initProxy()
	r.initReflect()

	r.initErrors()

	r.global.Eval = r.newNativeFunc(r.builtin_eval, nil, "eval", nil, 1)
	r.addToGlobal("eval", r.global.Eval)

	r.initMath()
	r.initJSON()

	r.initTypedArrays()
	r.initSymbol()
	r.initWeakSet()
	r.initWeakMap()
	r.initMap()
	r.initSet()
	r.initPromise()

	// 在 js 标准之外定义的额外函数库
	r.initHttp()
	r.initMySQL()
	r.initRedis()
	r.initEtcd()
	r.initFile()

	r.global.thrower = r.newNativeFunc(r.builtin_thrower, nil, "", nil, 0)
	r.global.throwerProperty = &valueProperty{
		getterFunc: r.global.thrower,
		setterFunc: r.global.thrower,
		accessor:   true,
	}
	r.object_freeze(FunctionCall{Arguments: []Value{r.global.thrower}})

	funcProtoObj._put("caller", &valueProperty{
		getterFunc:   r.global.thrower,
		setterFunc:   r.global.thrower,
		accessor:     true,
		configurable: true,
	})
	funcProtoObj._put("arguments", &valueProperty{
		getterFunc:   r.global.thrower,
		setterFunc:   r.global.thrower,
		accessor:     true,
		configurable: true,
	})
}

func (r *Runtime) typeErrorResult(throw bool, args ...any) {
	if throw {
		panic(r.NewTypeError(args...))
	}
}

func (r *Runtime) newError(typ *Object, format string, args ...any) Value {
	var msg string
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	} else {
		msg = format
	}
	return r.builtin_new(typ, []Value{newStringValue(msg)})
}

func (r *Runtime) throwReferenceError(name unistring.String) {
	panic(r.newError(r.global.ReferenceError, "%s is not defined", name))
}

func (r *Runtime) newSyntaxError(msg string, offset int) Value {
	return r.builtin_new(r.global.SyntaxError, []Value{newStringValue(msg)})
}

func newBaseObjectObj(obj, proto *Object, class string) *baseObject {
	o := &baseObject{
		class:      class,
		val:        obj,
		extensible: true,
		prototype:  proto,
	}
	obj.self = o
	o.init()
	return o
}

func newGuardedObj(proto *Object, class string) *guardedObject {
	return &guardedObject{
		baseObject: baseObject{
			class:      class,
			extensible: true,
			prototype:  proto,
		},
	}
}

func (r *Runtime) newBaseObject(proto *Object, class string) (o *baseObject) {
	v := &Object{runtime: r}
	return newBaseObjectObj(v, proto, class)
}

func (r *Runtime) newGuardedObject(proto *Object, class string) (o *guardedObject) {
	v := &Object{runtime: r}
	o = newGuardedObj(proto, class)
	v.self = o
	o.val = v
	o.init()
	return
}

func (r *Runtime) NewObject() (v *Object) {
	return r.newBaseObject(r.global.ObjectPrototype, classObject).val
}

func (r *Runtime) CreateObject(proto *Object) *Object {
	return r.newBaseObject(proto, classObject).val
}

func (r *Runtime) NewArray(items ...any) *Object {
	values := make([]Value, len(items))
	for i, item := range items {
		values[i] = r.ToValue(item)
	}
	return r.newArrayValues(values)
}

func (r *Runtime) NewTypeError(args ...any) *Object {
	msg := ""
	if len(args) > 0 {
		f, _ := args[0].(string)
		msg = fmt.Sprintf(f, args[1:]...)
	}
	return r.builtin_new(r.global.TypeError, []Value{newStringValue(msg)})
}

func (r *Runtime) NewGoError(err error) *Object {
	e := r.newError(r.global.GoError, err.Error()).(*Object)
	_ = e.Set("value", err)
	return e
}

func (r *Runtime) newFunc(name unistring.String, length int, strict bool) (f *funcObject) {
	v := &Object{runtime: r}

	f = &funcObject{}
	f.class = classFunction
	f.val = v
	f.extensible = true
	f.strict = strict
	v.self = f
	f.prototype = r.global.FunctionPrototype
	f.init(name, intToValue(int64(length)))
	return
}

func (r *Runtime) newClassFunc(name unistring.String, length int, proto *Object, derived bool) (f *classFuncObject) {
	v := &Object{runtime: r}

	f = &classFuncObject{}
	f.class = classFunction
	f.val = v
	f.extensible = true
	f.strict = true
	f.derived = derived
	v.self = f
	f.prototype = proto
	f.init(name, intToValue(int64(length)))
	return
}

func (r *Runtime) newMethod(name unistring.String, length int, strict bool) (f *methodFuncObject) {
	v := &Object{runtime: r}

	f = &methodFuncObject{}
	f.class = classFunction
	f.val = v
	f.extensible = true
	f.strict = strict
	v.self = f
	f.prototype = r.global.FunctionPrototype
	f.init(name, intToValue(int64(length)))
	return
}

func (r *Runtime) newArrowFunc(name unistring.String, length int, strict bool) (f *arrowFuncObject) {
	v := &Object{runtime: r}

	f = &arrowFuncObject{}
	f.class = classFunction
	f.val = v
	f.extensible = true
	f.strict = strict

	vm := r.vm

	f.newTarget = vm.newTarget
	v.self = f
	f.prototype = r.global.FunctionPrototype
	f.init(name, intToValue(int64(length)))
	return
}

func (r *Runtime) newNativeFuncObj(v *Object, call func(FunctionCall) Value, construct func(args []Value, proto *Object) *Object, name unistring.String, proto *Object, length Value) *nativeFuncObject {
	f := &nativeFuncObject{
		baseFuncObject: baseFuncObject{
			baseObject: baseObject{
				class:      classFunction,
				val:        v,
				extensible: true,
				prototype:  r.global.FunctionPrototype,
			},
		},
		f:         call,
		construct: r.wrapNativeConstruct(construct, proto),
	}
	v.self = f
	f.init(name, length)
	if proto != nil {
		f._putProp("prototype", proto, false, false, false)
	}
	return f
}

func (r *Runtime) newNativeConstructor(call func(ConstructorCall) *Object, name unistring.String, length int64) *Object {
	v := &Object{runtime: r}

	f := &nativeFuncObject{
		baseFuncObject: baseFuncObject{
			baseObject: baseObject{
				class:      classFunction,
				val:        v,
				extensible: true,
				prototype:  r.global.FunctionPrototype,
			},
		},
	}

	f.f = func(c FunctionCall) Value {
		thisObj, _ := c.This.(*Object)
		if thisObj != nil {
			res := call(ConstructorCall{
				This:      thisObj,
				Arguments: c.Arguments,
			})
			if res == nil {
				return _undefined
			}
			return res
		}
		return f.defaultConstruct(call, c.Arguments, nil)
	}

	f.construct = func(args []Value, newTarget *Object) *Object {
		return f.defaultConstruct(call, args, newTarget)
	}

	v.self = f
	f.init(name, intToValue(length))

	proto := r.NewObject()
	proto.self._putProp("constructor", v, true, false, true)
	f._putProp("prototype", proto, true, false, false)

	return v
}

func (r *Runtime) newNativeConstructOnly(v *Object, ctor func(args []Value, newTarget *Object) *Object, defaultProto *Object, name unistring.String, length int64) *nativeFuncObject {
	return r.newNativeFuncAndConstruct(v, func(call FunctionCall) Value {
		return ctor(call.Arguments, nil)
	},
		func(args []Value, newTarget *Object) *Object {
			if newTarget == nil {
				newTarget = v
			}
			return ctor(args, newTarget)
		}, defaultProto, name, intToValue(length))
}

func (r *Runtime) newNativeFuncAndConstruct(v *Object, call func(call FunctionCall) Value, ctor func(args []Value, newTarget *Object) *Object, defaultProto *Object, name unistring.String, l Value) *nativeFuncObject {
	if v == nil {
		v = &Object{runtime: r}
	}

	f := &nativeFuncObject{
		baseFuncObject: baseFuncObject{
			baseObject: baseObject{
				class:      classFunction,
				val:        v,
				extensible: true,
				prototype:  r.global.FunctionPrototype,
			},
		},
		f:         call,
		construct: ctor,
	}
	v.self = f
	f.init(name, l)
	if defaultProto != nil {
		f._putProp("prototype", defaultProto, false, false, false)
	}

	return f
}

func (r *Runtime) newNativeFunc(call func(FunctionCall) Value, construct func(args []Value, proto *Object) *Object, name unistring.String, proto *Object, length int) *Object {
	v := &Object{runtime: r}

	f := &nativeFuncObject{
		baseFuncObject: baseFuncObject{
			baseObject: baseObject{
				class:      classFunction,
				val:        v,
				extensible: true,
				prototype:  r.global.FunctionPrototype,
			},
		},
		f:         call,
		construct: r.wrapNativeConstruct(construct, proto),
	}
	v.self = f
	f.init(name, intToValue(int64(length)))
	if proto != nil {
		f._putProp("prototype", proto, false, false, false)
		proto.self._putProp("constructor", v, true, false, true)
	}
	return v
}

func (r *Runtime) newNativeFuncConstructObj(v *Object, construct func(args []Value, proto *Object) *Object, name unistring.String, proto *Object, length int) *nativeFuncObject {
	f := &nativeFuncObject{
		baseFuncObject: baseFuncObject{
			baseObject: baseObject{
				class:      classFunction,
				val:        v,
				extensible: true,
				prototype:  r.global.FunctionPrototype,
			},
		},
		f:         r.constructToCall(construct, proto),
		construct: r.wrapNativeConstruct(construct, proto),
	}

	f.init(name, intToValue(int64(length)))
	if proto != nil {
		f._putProp("prototype", proto, false, false, false)
	}
	return f
}

func (r *Runtime) newNativeFuncConstruct(construct func(args []Value, proto *Object) *Object, name unistring.String, prototype *Object, length int64) *Object {
	return r.newNativeFuncConstructProto(construct, name, prototype, r.global.FunctionPrototype, length)
}

func (r *Runtime) newNativeFuncConstructProto(construct func(args []Value, proto *Object) *Object, name unistring.String, prototype, proto *Object, length int64) *Object {
	v := &Object{runtime: r}

	f := &nativeFuncObject{}
	f.class = classFunction
	f.val = v
	f.extensible = true
	v.self = f
	f.prototype = proto
	f.f = r.constructToCall(construct, prototype)
	f.construct = r.wrapNativeConstruct(construct, prototype)
	f.init(name, intToValue(length))
	if prototype != nil {
		f._putProp("prototype", prototype, false, false, false)
		prototype.self._putProp("constructor", v, true, false, true)
	}
	return v
}

func (r *Runtime) newPrimitiveObject(value Value, proto *Object, class string) *Object {
	v := &Object{runtime: r}

	o := &primitiveValueObject{}
	o.class = class
	o.val = v
	o.extensible = true
	v.self = o
	o.prototype = proto
	o.pValue = value
	o.init()
	return v
}

func (r *Runtime) builtin_Number(call FunctionCall) Value {
	if len(call.Arguments) > 0 {
		return call.Arguments[0].ToNumber()
	} else {
		return valueInt(0)
	}
}

func (r *Runtime) builtin_newNumber(args []Value, proto *Object) *Object {
	var v Value
	if len(args) > 0 {
		v = args[0].ToNumber()
	} else {
		v = intToValue(0)
	}
	return r.newPrimitiveObject(v, proto, classNumber)
}

func (r *Runtime) builtin_Boolean(call FunctionCall) Value {
	if len(call.Arguments) > 0 {
		if call.Arguments[0].ToBoolean() {
			return valueTrue
		} else {
			return valueFalse
		}
	} else {
		return valueFalse
	}
}

func (r *Runtime) builtin_newBoolean(args []Value, proto *Object) *Object {
	var v Value
	if len(args) > 0 {
		if args[0].ToBoolean() {
			v = valueTrue
		} else {
			v = valueFalse
		}
	} else {
		v = valueFalse
	}
	return r.newPrimitiveObject(v, proto, classBoolean)
}

func (r *Runtime) error_toString(call FunctionCall) Value {
	var nameStr, msgStr valueString
	obj := r.toObject(call.This)
	name := obj.self.getStr("name", nil)
	if name == nil || name == _undefined {
		nameStr = asciiString("Error")
	} else {
		nameStr = name.toString()
	}
	msg := obj.self.getStr("message", nil)
	if msg == nil || msg == _undefined {
		msgStr = stringEmpty
	} else {
		msgStr = msg.toString()
	}
	if nameStr.length() == 0 {
		return msgStr
	}
	if msgStr.length() == 0 {
		return nameStr
	}
	var sb valueStringBuilder
	sb.WriteString(nameStr)
	sb.WriteString(asciiString(": "))
	sb.WriteString(msgStr)
	return sb.String()
}

func (r *Runtime) builtin_new(construct *Object, args []Value) *Object {
	return r.toConstructor(construct)(args, nil)
}

func (r *Runtime) builtin_thrower(call FunctionCall) Value {
	obj := r.toObject(call.This)
	strict := true
	switch fn := obj.self.(type) {
	case *funcObject:
		strict = fn.strict
	}
	r.typeErrorResult(strict, "'caller', 'callee', and 'arguments' properties may not be accessed on strict mode functions or the arguments objects for calls to them")
	return nil
}

func (r *Runtime) eval(srcVal valueString, direct, strict bool) Value {
	src := escapeInvalidUtf16(srcVal)
	vm := r.vm
	inGlobal := true
	if direct {
		for s := vm.stash; s != nil; s = s.outer {
			if s.isVariable() {
				inGlobal = false
				break
			}
		}
	}
	vm.pushCtx()
	funcObj := _undefined
	if !direct {
		vm.stash = &r.global.stash
		vm.privEnv = nil
	} else {
		if sb := vm.sb; sb > 0 {
			funcObj = vm.stack[sb-1]
		}
	}
	p, err := r.compile("<eval>", src, strict, inGlobal, r.vm)
	if err != nil {
		panic(err)
	}

	vm.prg = p
	vm.pc = 0
	vm.args = 0
	vm.result = _undefined
	vm.push(funcObj)
	vm.sb = vm.sp
	vm.push(nil) // this
	vm.run()
	retval := vm.result
	vm.popCtx()
	vm.halt = false
	vm.sp -= 2
	return retval
}

func (r *Runtime) builtin_eval(call FunctionCall) Value {
	if len(call.Arguments) == 0 {
		return _undefined
	}
	if str, ok := call.Arguments[0].(valueString); ok {
		return r.eval(str, false, false)
	}
	return call.Arguments[0]
}

func (r *Runtime) constructToCall(construct func(args []Value, proto *Object) *Object, proto *Object) func(call FunctionCall) Value {
	return func(call FunctionCall) Value {
		return construct(call.Arguments, proto)
	}
}

func (r *Runtime) wrapNativeConstruct(c func(args []Value, proto *Object) *Object, proto *Object) func(args []Value, newTarget *Object) *Object {
	if c == nil {
		return nil
	}
	return func(args []Value, newTarget *Object) *Object {
		var p *Object
		if newTarget != nil {
			if pp, ok := newTarget.self.getStr("prototype", nil).(*Object); ok {
				p = pp
			}
		}
		if p == nil {
			p = proto
		}
		return c(args, p)
	}
}

func (r *Runtime) toCallable(v Value) func(FunctionCall) Value {
	if call, ok := r.toObject(v).self.assertCallable(); ok {
		return call
	}
	r.typeErrorResult(true, "Value is not callable: %s", v.toString())
	return nil
}

func (r *Runtime) checkObjectCoercible(v Value) {
	switch v.(type) {
	case valueUndefined, valueNull:
		r.typeErrorResult(true, "Value is not object coercible")
	}
}

func toInt8(v Value) int8 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return int8(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return int8(int64(f))
		}
	}
	return 0
}

func toUint8(v Value) uint8 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return uint8(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return uint8(int64(f))
		}
	}
	return 0
}

func toUint8Clamp(v Value) uint8 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		if i < 0 {
			return 0
		}
		if i <= 255 {
			return uint8(i)
		}
		return 255
	}

	if num, ok := v.(valueFloat); ok {
		num := float64(num)
		if !math.IsNaN(num) {
			if num < 0 {
				return 0
			}
			if num > 255 {
				return 255
			}
			f := math.Floor(num)
			f1 := f + 0.5
			if f1 < num {
				return uint8(f + 1)
			}
			if f1 > num {
				return uint8(f)
			}
			r := uint8(f)
			if r&1 != 0 {
				return r + 1
			}
			return r
		}
	}
	return 0
}

func toInt16(v Value) int16 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return int16(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return int16(int64(f))
		}
	}
	return 0
}

func toUint16(v Value) uint16 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return uint16(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return uint16(int64(f))
		}
	}
	return 0
}

func toInt32(v Value) int32 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return int32(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return int32(int64(f))
		}
	}
	return 0
}

func toUint32(v Value) uint32 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return uint32(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return uint32(int64(f))
		}
	}
	return 0
}

func toInt64(v Value) int64 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return int64(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return int64(f)
		}
	}
	return 0
}

func toUint64(v Value) uint64 {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return uint64(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return uint64(int64(f))
		}
	}
	return 0
}

func toInt(v Value) int {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return int(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return int(f)
		}
	}
	return 0
}

func toUint(v Value) uint {
	v = v.ToNumber()
	if i, ok := v.(valueInt); ok {
		return uint(i)
	}

	if f, ok := v.(valueFloat); ok {
		f := float64(f)
		if !math.IsNaN(f) && !math.IsInf(f, 0) {
			return uint(int64(f))
		}
	}
	return 0
}

func toFloat32(v Value) float32 {
	return float32(v.ToFloat())
}

func toLength(v Value) int64 {
	if v == nil {
		return 0
	}
	i := v.ToInteger()
	if i < 0 {
		return 0
	}
	if i >= maxInt {
		return maxInt - 1
	}
	return i
}

func (r *Runtime) toLengthUint32(v Value) uint32 {
	var intVal int64
repeat:
	switch num := v.(type) {
	case valueInt:
		intVal = int64(num)
	case valueFloat:
		if v != _negativeZero {
			if i, ok := floatToInt(float64(num)); ok {
				intVal = i
			} else {
				goto fail
			}
		}
	case valueString:
		v = num.ToNumber()
		goto repeat
	default:
		// 必须的遗留行为 https://tc39.es/ecma262/#sec-arraysetlength
		n2 := toUint32(v)
		n1 := v.ToNumber()
		if f, ok := n1.(valueFloat); ok {
			f := float64(f)
			if f != 0 || !math.Signbit(f) {
				goto fail
			}
		}
		if n1.ToInteger() != int64(n2) {
			goto fail
		}
		return n2
	}
	if intVal >= 0 && intVal <= math.MaxUint32 {
		return uint32(intVal)
	}
fail:
	panic(r.newError(r.global.RangeError, "Invalid array length"))
}

func toIntStrict(i int64) int {
	if bits.UintSize == 32 {
		if i > math.MaxInt32 || i < math.MinInt32 {
			panic(rangeError("Integer value overflows 32-bit int"))
		}
	}
	return int(i)
}

func toIntClamp(i int64) int {
	if bits.UintSize == 32 {
		if i > math.MaxInt32 {
			return math.MaxInt32
		}
		if i < math.MinInt32 {
			return math.MinInt32
		}
	}
	return int(i)
}

func (r *Runtime) toIndex(v Value) int {
	num := v.ToInteger()
	if num >= 0 && num < maxInt {
		if bits.UintSize == 32 && num >= math.MaxInt32 {
			panic(r.newError(r.global.RangeError, "Index %s overflows int", v.String()))
		}
		return int(num)
	}
	panic(r.newError(r.global.RangeError, "Invalid index %s", v.String()))
}

func (r *Runtime) toBoolean(b bool) Value {
	if b {
		return valueTrue
	} else {
		return valueFalse
	}
}

// New 创建一个 Javascript 运行时的实例，可以用来运行代码。多个实例可以被创建并同时使用，但是不能在不同的运行时之间传递 JS 值
func New() *Runtime {
	r := &Runtime{}
	r.init()
	return r
}

// Compile 将会创建一个 Javascript 代码的内部形式，之后可以使用 Runtime.RunProgram() 方法运行
// 这个内部形式不会以任何方式链接到运行时，因此可以在多个运行时中运行
func Compile(name, src string, strict bool) (*Program, error) {
	return compile(name, src, strict, true, nil)
}

// CompileAST 与 Compile 一致，但是它接受一个 Javascript AST 形式的代码
func CompileAST(prg *jast.Program, strict bool) (*Program, error) {
	return compileAST(prg, strict, true, nil)
}

// MustCompile 与 Compile 一样，但是如果代码不能被编译，就会出现 panic
// 它简化了持有已编译的 Javascript 代码的全局变量的安全初始化
func MustCompile(name, src string, strict bool) *Program {
	prg, err := Compile(name, src, strict)
	if err != nil {
		panic(err)
	}

	return prg
}

// Parse 接收一个源字符串并产生一个解析的 AST。如果你想向解析器传递选项，请使用此函数
// 例如：
//	 p, err := Parse("test.js", "var a = true", parser.WithDisableSourceMaps)
//	 if err != nil { /* ... */ }
//	 prg, err := CompileAST(p, true)
//	 // ...
//
// 否则就使用 Compile，它结合了这两个步骤
func Parse(name, src string, options ...parser.Option) (prg *jast.Program, err error) {
	prg, err1 := parser.ParseFile(nil, name, src, 0, options...)
	if err1 != nil {
		err = &CompilerSyntaxError{
			CompilerError: CompilerError{
				Message: err1.Error(),
			},
		}
	}
	return
}

func compile(name, src string, strict, inGlobal bool, evalVm *vm, parserOptions ...parser.Option) (p *Program, err error) {
	prg, err := Parse(name, src, parserOptions...)
	if err != nil {
		return
	}

	return compileAST(prg, strict, inGlobal, evalVm)
}

func compileAST(prg *jast.Program, strict, inGlobal bool, evalVm *vm) (p *Program, err error) {
	c := newCompiler()

	defer func() {
		if x := recover(); x != nil {
			p = nil
			switch x1 := x.(type) {
			case *CompilerSyntaxError:
				err = x1
			default:
				panic(x)
			}
		}
	}()

	c.compile(prg, strict, inGlobal, evalVm)
	p = c.p
	return
}

func (r *Runtime) compile(name, src string, strict, inGlobal bool, evalVm *vm) (p *Program, err error) {
	p, err = compile(name, src, strict, inGlobal, evalVm, r.parserOptions...)
	if err != nil {
		switch x1 := err.(type) {
		case *CompilerSyntaxError:
			err = &Exception{
				val: r.builtin_new(r.global.SyntaxError, []Value{newStringValue(x1.Error())}),
			}
		case *CompilerReferenceError:
			err = &Exception{
				val: r.newError(r.global.ReferenceError, x1.Message),
			}
		}
	}
	return
}

// RunString 在全局环境中执行给定的字符串
func (r *Runtime) RunString(str string) (Value, error) {
	return r.RunScript("", str)
}

// RunScript 在全局环境中执行给定的字符串
func (r *Runtime) RunScript(name, src string) (Value, error) {
	p, err := r.compile(name, src, false, true, nil)

	if err != nil {
		return nil, err
	}

	return r.RunProgram(p)
}

// RunProgram 在全局上下文中执行预先编译的代码
func (r *Runtime) RunProgram(p *Program) (result Value, err error) {
	defer func() {
		if x := recover(); x != nil {
			if ex, ok := x.(*uncatchableException); ok {
				err = ex.err
				if len(r.vm.callStack) == 0 {
					r.leaveAbrupt()
				}
			} else {
				panic(x)
			}
		}
	}()
	vm := r.vm
	recursive := false
	if len(vm.callStack) > 0 {
		recursive = true
		vm.pushCtx()
		vm.stash = &r.global.stash
		vm.sb = vm.sp - 1
	}
	vm.prg = p
	vm.pc = 0
	vm.result = _undefined
	ex := vm.runTry()
	if ex == nil {
		result = r.vm.result
	} else {
		err = ex
	}
	if recursive {
		vm.popCtx()
		vm.halt = false
		vm.clearStack()
	} else {
		vm.stack = nil
		vm.prg = nil
		vm.funcName = ""
		r.leave()
	}
	return
}

// CaptureCallStack 将当前的调用堆栈帧附加到堆栈切片（可以是nil），直到指定的深度
// 如果深度 <= 0 或超过可用的帧数，则返回整个栈
// 这个方法对于并发使用并不安全，只能由运行中的脚本中的 Go 函数调用
func (r *Runtime) CaptureCallStack(depth int, stack []StackFrame) []StackFrame {
	l := len(r.vm.callStack)
	var offset int
	if depth > 0 {
		offset = l - depth + 1
		if offset < 0 {
			offset = 0
		}
	}
	if stack == nil {
		stack = make([]StackFrame, 0, l-offset+1)
	}
	return r.vm.captureStack(stack, offset)
}

// Interrupt 中断一个正在运行的 Javascript。相应的Go调用将返回一个包含 v 的 *InterruptedError
// 如果中断后的堆栈是空的，那么当前排队的 Promise 解析/拒绝作业将被清除而不被执行
// 注意，它只在 Javascript 代码中工作，它不会中断原生 Go 函数（包括所有的内置函数）
// 如果执行中断时，当前没有脚本正在运行，那么在下一次 Run*() 调用时就会立即中断，为了避免这种情况，请使用ClearInterrupt()
func (r *Runtime) Interrupt(v any) {
	r.vm.Interrupt(v)
}

// ClearInterrupt 重置中断标志。通常情况下，如果运行时有可能被 Interrupt() 所打断，则需要在运行时可以重新使用之前调用
func (r *Runtime) ClearInterrupt() {
	r.vm.ClearInterrupt()
}

/*
ToValue 将 Go 值转换为最合适类型的 Javascript值。结构类型（如 struct、map 和 slice）被包装起来，这样所产生的变化就会反映在原始值上，可以用 Value.Export() 来检索

警告! 这些包装好的 Go 值与原生 ECMAScript 值的行为方式不同。如果你打算在ECMAScript中修改它们，请记住以下注意事项：

1. 如果一个普通的 Javascript 对象被分配为包装的 Go struct、map 或数组中的一个元素，它将被 Export()，因此会被复制，这可能会导致JavaScript中出现意外的行为:
	m := map[string]any{}
	vm.Set("m", m)
	vm.RunString(`
	    var obj = {test: false};
	    m.obj = obj; // obj 被 Export()，即复制到一个新的 map[string]any 并且这个 map 被设置到 m["obj"]
	    obj.test = true; // 注意，这里的 m.obj.test 依然是 false
	`)
	fmt.Println(m["obj"].(map[string]any)["test"]) // 打印出 false

2. 如果你在 ECMAScript 中修改嵌套的非指针式复合类型（struct、slice 和数组），要小心，尽可能避免在 ECMAScript 中修改它们，ECMAScript 和 Go 的一个根本区别在于
   前者的所有对象都是引用，而在 Go 中你可以有一个字面的 struct 或数组。请看下面的例子:

	type S struct {
	    Field int
	}

	a := []S{{1}, {2}} // 节片的字面结构
	vm.Set("a", &a)
	vm.RunString(`
	    let tmp = {Field: 1};
	    a[0] = tmp;
	    a[1] = tmp;
	    tmp.Field = 2;
	`)

   在ECMAScript中，我们希望a[0].Field和a[1].Field等于2，但是这是不可能的(或者至少在没有复杂的引用跟踪的情况下是不可行的)。
   为了涵盖最常见的使用情况并避免过多的内存分配，"变化时复制" 的机制已被实现（对数组和struct）:

   * 当一个嵌套的复合值被访问时，返回的 ES 值成为对字面值的引用。这保证了像 'a[0].Field = 1' 这样的事情能如期进行，对 'a[0].Field' 的简单访问不会导致导致对 a[0] 的复制
   * 原始的容器（在上述例子中是 'a' ）会跟踪返回的引用值，如果 a[0] 被重新赋值（例如，通过直接赋值、删除或缩小数组），旧的 a[0] 被复制，先前的返回值成为该副本的一个引用

   例如：

	let tmp = a[0];                      // 没有复制，tmp 是对 a[0] 的引用
	tmp.Field = 1;                       // 此时 a[0].Field === 1
	a[0] = {Field: 2};                   // tmp现在是对旧值副本的引用（Field ===1）
	a[0].Field === 2 && tmp.Field === 1; // true

   * 由原地排序（使用 Array.prototype.sort() ）引起的数组值交换不被视为重新分配，而是引用被调整为指向新的索引
   * 对内部复合值的赋值总是进行复制（有时还进行类型转换）

	a[1] = tmp;    // a[1] 现在是 tmp 的复制
	tmp.Field = 3; // 不影响 a[1].Field

3. 非可寻址 struct、slice 和数组被复制。这有时可能会导致混乱，因为分配给内部字段的赋值似乎并不奏效

	a1 := []any{S{1}, S{2}}
	vm.Set("a1", &a1)
	vm.RunString(`
	   a1[0].Field === 1; // true
	   a1[0].Field = 2;
	   a1[0].Field === 2; // false, 因为它真正做的是复制 a1[0]，将其字段设置为2，并立即将其删除
	`)

   另一种方法是让 a1[0].Field 成为一个不可写的属性，如果需要修改的话，就需要手动复制值，但是这可能是不切实际的
   注意，这同样适用于 slice。如果一个 slice 是通过值传递的（而不是作为一个指针），调整 slice 的大小并不反映在原来的值。此外，扩展 slice 可能会导致底层数组被重新分配和复制
   例如:

	a := []any{1}
	vm.Set("a", a)
	vm.RunString(`a.push(2); a[0] = 0;`)
	fmt.Println(a[0]) // 打印 "1"

   关于个别类型的说明:
   #原始类型
      原始类型（数字、字符串、布尔）被转换为相应的 Javascript 原语
   # 字符串
      由于 ECMAScript（使用 UTF-16）和 Go（使用 UTF-8）从 JS 到 Go 的转换可能是有损失的。字符串值必须是一个有效的UTF-8。如果不是，无效的字符会被替换为utf8.RuneError
      但是后续的 Export() 的行为没有被指定（它可能返回原始值，或一个被替换的无效字符）
   # Nil
      Nil 被转换为 null
   # 函数
      func(FunctionCall) Value被视为一个本地 Javascript 函数。这将提高性能，因为没有自动转换参数和返回值的类型（这涉及到反射）。试图将该函数用作构造函数将导致 TypeError

      func(ConstructorCall) *Object 被视为一个本地构造函数，允许用 new 操作符一起使用

		func MyObject(call ConstructorCall) *Object {
			call.This.Set("method", method)
	        //...
	        // instance := &myCustomStruct{}
	        // instanceValue := vm.ToValue(instance).(*Object)
	        // instanceValue.SetPrototype(call.This.Prototype())
	        // return instanceValue
	        return nil
	     }
	     runtime.Set("MyObject", MyObject)

      那么它可以在 JS 中使用，如下：

        var o = new MyObject(arg);
	    var o1 = MyObject(arg); // 等价于上面的
	    o instanceof MyObject && o1 instanceof MyObject; // true

      当一个本地构造函数被直接调用时（没有new操作符），其行为取决于这个值：如果它是一个对象，它会被传递，否则会创建一个新的对象，就像它是用 new 操作符调用的。在这两种情况下，call.NewTarget 将是 nil

	  func(ConstructorCall, *Runtime) *Object 的处理方法同上，只是 *Runtime 也被作为参数传递
      任何其他的 Go 函数都会被包装起来，这样参数就会自动转换为所需的 Go 类型，而返回值被转换为 Javascript 值（使用此方法）。 如果无法转换，则会抛出 TypeError

      有多个返回值的函数返回一个数组。如果最后一个返回值是一个 error，它不会被返回，而是被转换成一个 JS 异常。如果错误是 *Exception，它将被原样抛出，否则它将被包裹在一个 GoError 中
      注意，如果正好有两个返回值，并且最后一个是 error，函数会原样返回第一个值，而不是一个数组

   # Structs
      struct 被转换为类似对象的值。字段和方法可以作为属性使用，它们的值是这个方法（ToValue()）的结果应用于相应的 Go 值
      字段属性是可写的、不可配置的，方法属性是不可写和不可配置的
      试图定义一个新的属性或删除一个现有的属性将会失败（在严格模式下抛出），除非它是一个 Symbol 属性。符号属性只存在于包装器中，不影响底层的 Go 值
      请注意，由于每次访问一个属性都会创建一个包装器，因此可能会导致一些意想不到的结果，例如：

		type Field struct{
		}
		type S struct {
			Field *Field
		}
		var s = S{
			Field: &Field{},
	 	}
	 	vm := New()
	 	vm.Set("s", &s)
	 	res, err := vm.RunString(`
	 		var sym = Symbol(66);
	 		var field1 = s.Field;
	 		field1[sym] = true;
	 		var field2 = s.Field;
	 		field1 === field2; // true, 因为 == 操作比较的是被包装的值，而不是包装器
	 		field1[sym] === true; // true
	 		field2[sym] === undefined; // true
	 	`)

      这同样适用于来自 map 和 slice 的值

   # 对 time.Time 的处理
      time.Time 没有得到特殊的处理，因此它的转换就像其他的结构一样，提供对其所有方法的访问
      这样做是故意的，而不是将其转换为 Date，因为这两种类型并不完全兼容，time.Time 包含时区，而 JS 的 Date 不包含，因此隐含地进行转换会导致信息的丢失
      如果你需要将其转换为 Date，可以在 JS 中完成。

		var d = new Date(goval.UnixNano()/1e6);

      或者在 Go 中完成：

	 	now := time.Now()
	 	vm := New()
	 	val, err := vm.New(vm.Get("Date").ToObject(vm), vm.ToValue(now.UnixNano()/1e6))
	 	if err != nil {
			...
	 	}
	 	vm.Set("d", val)

      请注意，Value.Export() 对于一个 Date 值会返回包含当地时区的 time.Time

   # Maps
      带有字符串或整数键类型的 map 被转换为 host 对象，其行为大体上与 Javascript 对象类似

   # 带有方法的 Maps
      如果一个 map 类型定义了方法，那么产生的 Object 的属性代表了它的方法，而不是通过 map 的 key
      这是因为在 Javascript中，object.key 和 object[key] 之间是没有区别的，这一点与 Go 不同
      如果需要访问 map 的值，可以通过定义另一个方法来实现，也可以通过定义一个外部 getter 函数

   # Slices
      slice 被转换为 host 对象，其行为在很大程度上类似于 Javascript 的数组。它有适当的原型，所有常用的方法都可以使用
      然而，有一点需要注意：转换后的数组不能包含空洞(因为 Go slice 不能)。这意味着 hasOwnProperty(n) 总是在 n < length 时返回 true
      删除一个 索引小于 length 的项将被设置为零值（但属性会保留）。nil slice 元素将被转换为 null
      访问一个超过 length 的元素会返回 undefined。也请看上面的警告，关于将 slice 作为值（相较于指针）

   # 数组
      数组的转换与 slice 的转换类似，只是产生的数组不能调整大小（length 属性是不可写的）
      任何其他类型被转换为基于反射的通用 host 对象。根据底层类型的不同，它的行为类似于与数字、字符串、布尔值或对象
      请注意，底层类型不会丢失，调用 Export() 返回原始的 Go 值。这适用于所有基于反射的类型
*/
func (r *Runtime) ToValue(i any) Value {
	switch i := i.(type) {
	case nil:
		return _null
	case *Object:
		if i == nil || i.self == nil {
			return _null
		}
		if i.runtime != nil && i.runtime != r {
			panic(r.NewTypeError("Illegal runtime transition of an Object"))
		}
		return i
	case valueContainer:
		return i.toValue(r)
	case Value:
		return i
	case string:
		if len(i) <= 16 {
			if u := unistring.Scan(i); u != nil {
				return &importedString{s: i, u: u, scanned: true}
			}
			return asciiString(i)
		}
		return &importedString{s: i}
	case bool:
		if i {
			return valueTrue
		} else {
			return valueFalse
		}
	case func(FunctionCall) Value:
		name := unistring.NewFromString(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name())
		return r.newNativeFunc(i, nil, name, nil, 0)
	case func(FunctionCall, *Runtime) Value:
		name := unistring.NewFromString(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name())
		return r.newNativeFunc(func(call FunctionCall) Value {
			return i(call, r)
		}, nil, name, nil, 0)
	case func(ConstructorCall) *Object:
		name := unistring.NewFromString(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name())
		return r.newNativeConstructor(i, name, 0)
	case func(ConstructorCall, *Runtime) *Object:
		name := unistring.NewFromString(runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name())
		return r.newNativeConstructor(func(call ConstructorCall) *Object {
			return i(call, r)
		}, name, 0)
	case int:
		return intToValue(int64(i))
	case int8:
		return intToValue(int64(i))
	case int16:
		return intToValue(int64(i))
	case int32:
		return intToValue(int64(i))
	case int64:
		return intToValue(i)
	case uint:
		if uint64(i) <= math.MaxInt64 {
			return intToValue(int64(i))
		} else {
			return floatToValue(float64(i))
		}
	case uint8:
		return intToValue(int64(i))
	case uint16:
		return intToValue(int64(i))
	case uint32:
		return intToValue(int64(i))
	case uint64:
		if i <= math.MaxInt64 {
			return intToValue(int64(i))
		}
		return floatToValue(float64(i))
	case float32:
		return floatToValue(float64(i))
	case float64:
		return floatToValue(i)
	case map[string]any:
		if i == nil {
			return _null
		}
		obj := &Object{runtime: r}
		m := &objectGoMapSimple{
			baseObject: baseObject{
				val:        obj,
				extensible: true,
			},
			data: i,
		}
		obj.self = m
		m.init()
		return obj
	case []any:
		if i == nil {
			return _null
		}
		return r.newObjectGoSlice(&i).val
	case *[]any:
		if i == nil {
			return _null
		}
		return r.newObjectGoSlice(i).val
	}

	return r.reflectValueToValue(reflect.ValueOf(i))
}

func (r *Runtime) reflectValueToValue(origValue reflect.Value) Value {
	value := origValue
	for value.Kind() == reflect.Ptr {
		value = reflect.Indirect(value)
	}

	if !value.IsValid() {
		return _null
	}

	switch value.Kind() {
	case reflect.Map:
		if value.Type().NumMethod() == 0 {
			switch value.Type().Key().Kind() {
			case reflect.String, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
				reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
				reflect.Float64, reflect.Float32:

				obj := &Object{runtime: r}
				m := &objectGoMapReflect{
					objectGoReflect: objectGoReflect{
						baseObject: baseObject{
							val:        obj,
							extensible: true,
						},
						origValue: origValue,
						value:     value,
					},
				}
				m.init()
				obj.self = m
				return obj
			}
		}
	case reflect.Array:
		obj := &Object{runtime: r}
		a := &objectGoArrayReflect{
			objectGoReflect: objectGoReflect{
				baseObject: baseObject{
					val: obj,
				},
				origValue: origValue,
				value:     value,
			},
		}
		a.init()
		obj.self = a
		return obj
	case reflect.Slice:
		obj := &Object{runtime: r}
		a := &objectGoSliceReflect{
			objectGoArrayReflect: objectGoArrayReflect{
				objectGoReflect: objectGoReflect{
					baseObject: baseObject{
						val: obj,
					},
					origValue: origValue,
					value:     value,
				},
			},
		}
		a.init()
		obj.self = a
		return obj
	case reflect.Func:
		name := unistring.NewFromString(runtime.FuncForPC(value.Pointer()).Name())
		return r.newNativeFunc(r.wrapReflectFunc(value), nil, name, nil, value.Type().NumIn())
	}

	obj := &Object{runtime: r}
	o := &objectGoReflect{
		baseObject: baseObject{
			val: obj,
		},
		origValue: origValue,
		value:     value,
	}
	obj.self = o
	o.init()
	return obj
}

func (r *Runtime) wrapReflectFunc(value reflect.Value) func(FunctionCall) Value {
	return func(call FunctionCall) Value {
		typ := value.Type()
		nargs := typ.NumIn()
		var in []reflect.Value

		if l := len(call.Arguments); l < nargs {
			// 用零值填补缺失的参数
			n := nargs
			if typ.IsVariadic() {
				n--
			}
			in = make([]reflect.Value, n)
			for i := l; i < n; i++ {
				in[i] = reflect.Zero(typ.In(i))
			}
		} else {
			if l > nargs && !typ.IsVariadic() {
				l = nargs
			}
			in = make([]reflect.Value, l)
		}

		for i, a := range call.Arguments {
			var t reflect.Type

			n := i
			if n >= nargs-1 && typ.IsVariadic() {
				if n > nargs-1 {
					n = nargs - 1
				}

				t = typ.In(n).Elem()
			} else if n > nargs-1 {
				break
			} else {
				t = typ.In(n)
			}

			v := reflect.New(t).Elem()
			err := r.toReflectValue(a, v, &objectExportCtx{})
			if err != nil {
				panic(r.NewTypeError("could not convert function call parameter %d: %v", i, err))
			}
			in[i] = v
		}

		out := value.Call(in)
		if len(out) == 0 {
			return _undefined
		}

		if last := out[len(out)-1]; last.Type().Name() == "error" {
			if !last.IsNil() {
				err := last.Interface()
				if _, ok := err.(*Exception); ok {
					panic(err)
				}
				panic(r.NewGoError(last.Interface().(error)))
			}
			out = out[:len(out)-1]
		}

		switch len(out) {
		case 0:
			return _undefined
		case 1:
			return r.ToValue(out[0].Interface())
		default:
			s := make([]any, len(out))
			for i, v := range out {
				s[i] = v.Interface()
			}

			return r.ToValue(s)
		}
	}
}

func (r *Runtime) toReflectValue(v Value, dst reflect.Value, ctx *objectExportCtx) error {
	typ := dst.Type()

	if typ == typeValue {
		dst.Set(reflect.ValueOf(v))
		return nil
	}

	if typ == typeObject {
		if obj, ok := v.(*Object); ok {
			dst.Set(reflect.ValueOf(obj))
			return nil
		}
	}

	if typ == typeCallable {
		if fn, ok := AssertFunction(v); ok {
			dst.Set(reflect.ValueOf(fn))
			return nil
		}
	}

	et := v.ExportType()
	if et == nil || et == reflectTypeNil {
		dst.Set(reflect.Zero(typ))
		return nil
	}

	kind := typ.Kind()
	for i := 0; ; i++ {
		if et.AssignableTo(typ) {
			ev := reflect.ValueOf(exportValue(v, ctx))
			for ; i > 0; i-- {
				ev = ev.Elem()
			}
			dst.Set(ev)
			return nil
		}
		expKind := et.Kind()
		if expKind == kind && et.ConvertibleTo(typ) || expKind == reflect.String && typ == typeBytes {
			ev := reflect.ValueOf(exportValue(v, ctx))
			for ; i > 0; i-- {
				ev = ev.Elem()
			}
			dst.Set(ev.Convert(typ))
			return nil
		}
		if expKind == reflect.Ptr {
			et = et.Elem()
		} else {
			break
		}
	}

	if typ == typeTime {
		if obj, ok := v.(*Object); ok {
			if d, ok := obj.self.(*dateObject); ok {
				dst.Set(reflect.ValueOf(d.time()))
				return nil
			}
		}
		if et.Kind() == reflect.String {
			tme, ok := dateParse(v.String())
			if !ok {
				return fmt.Errorf("could not convert string %v to %v", v, typ)
			}
			dst.Set(reflect.ValueOf(tme))
			return nil
		}
	}

	switch kind {
	case reflect.String:
		dst.Set(reflect.ValueOf(v.String()).Convert(typ))
		return nil
	case reflect.Bool:
		dst.Set(reflect.ValueOf(v.ToBoolean()).Convert(typ))
		return nil
	case reflect.Int:
		dst.Set(reflect.ValueOf(toInt(v)).Convert(typ))
		return nil
	case reflect.Int64:
		dst.Set(reflect.ValueOf(toInt64(v)).Convert(typ))
		return nil
	case reflect.Int32:
		dst.Set(reflect.ValueOf(toInt32(v)).Convert(typ))
		return nil
	case reflect.Int16:
		dst.Set(reflect.ValueOf(toInt16(v)).Convert(typ))
		return nil
	case reflect.Int8:
		dst.Set(reflect.ValueOf(toInt8(v)).Convert(typ))
		return nil
	case reflect.Uint:
		dst.Set(reflect.ValueOf(toUint(v)).Convert(typ))
		return nil
	case reflect.Uint64:
		dst.Set(reflect.ValueOf(toUint64(v)).Convert(typ))
		return nil
	case reflect.Uint32:
		dst.Set(reflect.ValueOf(toUint32(v)).Convert(typ))
		return nil
	case reflect.Uint16:
		dst.Set(reflect.ValueOf(toUint16(v)).Convert(typ))
		return nil
	case reflect.Uint8:
		dst.Set(reflect.ValueOf(toUint8(v)).Convert(typ))
		return nil
	case reflect.Float64:
		dst.Set(reflect.ValueOf(v.ToFloat()).Convert(typ))
		return nil
	case reflect.Float32:
		dst.Set(reflect.ValueOf(toFloat32(v)).Convert(typ))
		return nil
	case reflect.Slice, reflect.Array:
		if o, ok := v.(*Object); ok {
			if v, exists := ctx.getTyped(o, typ); exists {
				dst.Set(reflect.ValueOf(v))
				return nil
			}
			return o.self.exportToArrayOrSlice(dst, typ, ctx)
		}
	case reflect.Map:
		if o, ok := v.(*Object); ok {
			if v, exists := ctx.getTyped(o, typ); exists {
				dst.Set(reflect.ValueOf(v))
				return nil
			}
			return o.self.exportToMap(dst, typ, ctx)
		}
	case reflect.Struct:
		if o, ok := v.(*Object); ok {
			t := reflect.PtrTo(typ)
			if v, exists := ctx.getTyped(o, t); exists {
				dst.Set(reflect.ValueOf(v).Elem())
				return nil
			}
			s := dst
			ctx.putTyped(o, t, s.Addr().Interface())
			for i := 0; i < typ.NumField(); i++ {
				field := typ.Field(i)
				if ast.IsExported(field.Name) {
					name := field.Name
					if r.fieldNameMapper != nil {
						name = r.fieldNameMapper.FieldName(typ, field)
					}
					var v Value
					if field.Anonymous {
						v = o
					} else {
						v = o.self.getStr(unistring.NewFromString(name), nil)
					}

					if v != nil {
						err := r.toReflectValue(v, s.Field(i), ctx)
						if err != nil {
							return fmt.Errorf("could not convert struct value %v to %v for field %s: %w", v, field.Type, field.Name, err)
						}
					}
				}
			}
			return nil
		}
	case reflect.Func:
		if fn, ok := AssertFunction(v); ok {
			dst.Set(reflect.MakeFunc(typ, r.wrapJSFunc(fn, typ)))
			return nil
		}
	case reflect.Ptr:
		if o, ok := v.(*Object); ok {
			if v, exists := ctx.getTyped(o, typ); exists {
				dst.Set(reflect.ValueOf(v))
				return nil
			}
		}
		if dst.IsNil() {
			dst.Set(reflect.New(typ.Elem()))
		}
		return r.toReflectValue(v, dst.Elem(), ctx)
	}

	return fmt.Errorf("could not convert %v to %v", v, typ)
}

func (r *Runtime) wrapJSFunc(fn Callable, typ reflect.Type) func(args []reflect.Value) (results []reflect.Value) {
	return func(args []reflect.Value) (results []reflect.Value) {
		jsArgs := make([]Value, len(args))
		for i, arg := range args {
			jsArgs[i] = r.ToValue(arg.Interface())
		}

		results = make([]reflect.Value, typ.NumOut())
		res, err := fn(_undefined, jsArgs...)
		if err == nil {
			if typ.NumOut() > 0 {
				v := reflect.New(typ.Out(0)).Elem()
				err = r.toReflectValue(res, v, &objectExportCtx{})
				if err == nil {
					results[0] = v
				}
			}
		}

		if err != nil {
			if typ.NumOut() == 2 && typ.Out(1).Name() == "error" {
				results[1] = reflect.ValueOf(err).Convert(typ.Out(1))
			} else {
				panic(err)
			}
		}

		for i, v := range results {
			if !v.IsValid() {
				results[i] = reflect.Zero(typ.Out(i))
			}
		}

		return
	}
}

/*
ExportTo 将一个 Javascript 值转换为指定的 Go 值。第二个参数必须是一个非空的指针，如果不能转换，则返回错误

   关于具体案例的说明:

   # 空接口
      导出为一个空的 any，与 Value.Export() 产生的类型相同的值

   # 数值类型
      导出到数字类型使用标准的 ECMAScript 转换操作，与向非钳制类型的数组项赋值时使用相同

   # 函数
      导出到一个 func 将创建一个严格类型的'网关'到一个可以从 Go 中调用的 ES 函数内
      使用 Runtime.ToValue() 将参数转换为 ES 值。如果 func 没有返回值，则返回值被忽略。如果 func 正好有一个返回值，
      则使用 ExportTo() 将其转换为适当的类型。如果 func 正好有 2 个返回值，并且第二个值是 error ，那么异常将被捕获并作为*Exception返回
      在所有其他情况下，异常会导致 panic。任何额外的返回值都被清零

      注意，如果你想捕捉并返回异常作为 error，并且你不需要返回值，func(...) error 将不能像预期那样工作。在这种情况下，error 被映射到函数的返回值，而不是异常，这仍然会导致 panic
      使用 func(...) (Value, error) 代替，并且忽略Value

      'this' 的值将永远被设置为 'undefined'

   # map 类型
      一个 ES map 可以被导出到 Go map 类型中。如果任何导出的键值是非哈希的，操作就会产生 panic（就像 reflect.Value.SetMapIndex() 那样）
      将一个 ES Set 导出到一个 map 类型，导致 map 被填充了（元素）->（零值）键/值对。如果有任何值是非哈希的，操作就会产生 panic（就像reflect.Value.SetMapIndex()那样）
      Symbol.iterator 被忽略了，任何其他对象都会用自己的可枚举的非符号属性来填充 map

   # slice 类型
      将一个 ES Set 导出到一个 slice 类型中，会导致其元素被导出
      将任何实现可迭代协议的对象导出为 slice 类型，都将导致 slice 被迭代的结果所填充
      数组被视为可迭代（即覆盖 Symbol.iterator 会影响结果）
      如果一个对象有 length 属性，并且不是一个函数，那么它就被当作数组一样处理。产生的 slice 将包含 obj[0], ... obj[length-1]
      对于任何其他对象，将返回一个错误

   # 数组类型
      只要长度匹配，任何可以导出为 slice 类型的东西也可以导出为数组类型。如果不匹配，就会返回一个错误

   # Proxy
      代理对象的处理方式与从 ES 代码中访问它们的属性（如 length 或 Symbol.iterator）相同。这意味着将它们导出到 Slice 类型中可以工作，
      但是将代理的 map 导出到 map 类型中不会同时导出其内容，因为代理不被认可为 map。这也适用于代理的 Set
*/
func (r *Runtime) ExportTo(v Value, target any) error {
	tval := reflect.ValueOf(target)
	if tval.Kind() != reflect.Ptr || tval.IsNil() {
		return errors.New("target must be a non-nil pointer")
	}
	return r.toReflectValue(v, tval.Elem(), &objectExportCtx{})
}

func (r *Runtime) GlobalObject() *Object {
	return r.globalObject
}

// Set 在全局环境中设置指定的变量
// 相当于在非严格模式下运行 "name = value"，需要首先使用 ToValue() 转换数值
func (r *Runtime) Set(name string, value any) error {
	return r.try(func() {
		name := unistring.NewFromString(name)
		v := r.ToValue(value)
		if ref := r.global.stash.getRefByName(name, false); ref != nil {
			ref.set(v)
		} else {
			r.globalObject.self.setOwnStr(name, v, true)
		}
	})
}

// Get 在全局环境中获取指定的变量
// 相当于在非严格模式下通过名称解除引用一个变量。如果变量未被定义，则返回nil
// 注意，这与 GlobalObject().Get(name) 不同，因为如果存在一个全局词法绑定（let 或 const），就会使用它
// 如果在这个过程中抛出了一个 Javascript 异常，这个方法将以 *Exception 的形式出现
func (r *Runtime) Get(name string) (ret Value) {
	r.tryPanic(func() {
		n := unistring.NewFromString(name)
		if v, exists := r.global.stash.getByName(n); exists {
			ret = v
		} else {
			ret = r.globalObject.self.getStr(n, nil)
		}
	})
	return
}

// SetRandSource 为这个 Runtime 设置随机源。如果不调用，则使用默认的 math/rand
func (r *Runtime) SetRandSource(source RandSource) {
	r.rand = source
}

// SetTimeSource 为这个Runtime设置当前的时间源，如果不调用，则使用默认的time.Now()
func (r *Runtime) SetTimeSource(now Now) {
	r.now = now
}

// SetParserOptions 设置解析器选项，以便在代码中被 RunString、RunScript 和 eval() 使用
func (r *Runtime) SetParserOptions(opts ...parser.Option) {
	r.parserOptions = opts
}

// SetMaxCallStackSize 设置最大的函数调用深度。当超过时，一个 *StackOverflowError 被抛出并由 RunProgram 或 Callable 调用返回
// 这对于防止无限递归引起的内存耗尽很有用。默认值是 math.MaxInt32
// 这个方法（和其他的 Set* 方法一样）对于并发使用是不安全的，只能从 vm 所在的协程或者当 vm 没有运行时调用
func (r *Runtime) SetMaxCallStackSize(size int) {
	r.vm.maxCallStackSize = size
}

// New 等同于 new 运算符，允许从 Go 中直接调用它
func (r *Runtime) New(construct Value, args ...Value) (o *Object, err error) {
	err = r.try(func() {
		o = r.builtin_new(r.toObject(construct), args)
	})
	return
}

// Callable 代表一个可以从 Go 中调用的 Javascript函数
type Callable func(this Value, args ...Value) (Value, error)

func AssertFunction(v Value) (Callable, bool) {
	if obj, ok := v.(*Object); ok {
		if f, ok := obj.self.assertCallable(); ok {
			return func(this Value, args ...Value) (ret Value, err error) {
				err = obj.runtime.runWrapped(func() {
					ret = f(FunctionCall{
						This:      this,
						Arguments: args,
					})
				})
				return
			}, true
		}
	}
	return nil, false
}

// Constructor 构造函数是一个可以用来调用构造函数的类型。第一个参数（newTarget）可以是 nil，将其设置为构造函数本身
type Constructor func(newTarget *Object, args ...Value) (*Object, error)

func AssertConstructor(v Value) (Constructor, bool) {
	if obj, ok := v.(*Object); ok {
		if ctor := obj.self.assertConstructor(); ctor != nil {
			return func(newTarget *Object, args ...Value) (ret *Object, err error) {
				err = obj.runtime.runWrapped(func() {
					ret = ctor(args, newTarget)
				})
				return
			}, true
		}
	}
	return nil, false
}

func (r *Runtime) runWrapped(f func()) (err error) {
	defer func() {
		if x := recover(); x != nil {
			if ex, ok := x.(*uncatchableException); ok {
				err = ex.err
				if len(r.vm.callStack) == 0 {
					r.leaveAbrupt()
				}
			} else {
				panic(x)
			}
		}
	}()
	ex := r.vm.try(f)
	if ex != nil {
		err = ex
	}
	r.vm.clearStack()
	if len(r.vm.callStack) == 0 {
		r.leave()
	}
	return
}

// IsUndefined 如果提供的值是未定义的，则返回 true
// 注意，它检查的是真正的未定义，而不是全局对象的'未定义'属性
func IsUndefined(v Value) bool {
	return v == _undefined
}

// IsNull 如果 Value 的值是 null，则返回 true
func IsNull(v Value) bool {
	return v == _null
}

// IsNaN 如果 Value 是 NaN，则返回 true
func IsNaN(v Value) bool {
	f, ok := v.(valueFloat)
	return ok && math.IsNaN(float64(f))
}

// IsInfinity 如果 Value 是正负无穷大，则返回 true
func IsInfinity(v Value) bool {
	return v == _positiveInf || v == _negativeInf
}

// Undefined 返回 JS 的未定义值。注意，如果全局的 "未定义" 属性被改变，它仍然返回原来的值
func Undefined() Value {
	return _undefined
}

// Null 返回 JS null 值
func Null() Value {
	return _null
}

// NaN 返回 JS NaN 值
func NaN() Value {
	return _NaN
}

// PositiveInf 返回 JS 正无穷大值
func PositiveInf() Value {
	return _positiveInf
}

// NegativeInf 返回 JS 负无穷大值
func NegativeInf() Value {
	return _negativeInf
}

func False() Value {
	return valueFalse
}

func True() Value {
	return valueTrue
}

func tryFunc(f func()) (ret any) {
	defer func() {
		ret = recover()
	}()

	f()
	return
}

func (r *Runtime) try(f func()) error {
	if ex := r.vm.try(f); ex != nil {
		return ex
	}
	return nil
}

func (r *Runtime) tryPanic(f func()) {
	if ex := r.vm.try(f); ex != nil {
		panic(ex)
	}
}

func (r *Runtime) toObject(v Value, args ...any) *Object {
	if obj, ok := v.(*Object); ok {
		return obj
	}
	if len(args) > 0 {
		panic(r.NewTypeError(args...))
	} else {
		var s string
		if v == nil {
			s = "undefined"
		} else {
			s = v.String()
		}
		panic(r.NewTypeError("Value is not an object: %s", s))
	}
}

func (r *Runtime) toNumber(v Value) Value {
	switch o := v.(type) {
	case valueInt, valueFloat:
		return v
	case *Object:
		if pvo, ok := o.self.(*primitiveValueObject); ok {
			return r.toNumber(pvo.pValue)
		}
	}
	panic(r.NewTypeError("Value is not a number: %s", v))
}

func (r *Runtime) speciesConstructor(o, defaultConstructor *Object) func(args []Value, newTarget *Object) *Object {
	c := o.self.getStr("constructor", nil)
	if c != nil && c != _undefined {
		c = r.toObject(c).self.getSym(SymSpecies, nil)
	}
	if c == nil || c == _undefined || c == _null {
		c = defaultConstructor
	}
	return r.toConstructor(c)
}

func (r *Runtime) speciesConstructorObj(o, defaultConstructor *Object) *Object {
	c := o.self.getStr("constructor", nil)
	if c != nil && c != _undefined {
		c = r.toObject(c).self.getSym(SymSpecies, nil)
	}
	if c == nil || c == _undefined || c == _null {
		return defaultConstructor
	}
	obj := r.toObject(c)
	if obj.self.assertConstructor() == nil {
		panic(r.NewTypeError("Value is not a constructor"))
	}
	return obj
}

func (r *Runtime) returnThis(call FunctionCall) Value {
	return call.This
}

func createDataProperty(o *Object, p Value, v Value) {
	o.defineOwnProperty(p, PropertyDescriptor{
		Writable:     FLAG_TRUE,
		Enumerable:   FLAG_TRUE,
		Configurable: FLAG_TRUE,
		Value:        v,
	}, false)
}

func createDataPropertyOrThrow(o *Object, p Value, v Value) {
	o.defineOwnProperty(p, PropertyDescriptor{
		Writable:     FLAG_TRUE,
		Enumerable:   FLAG_TRUE,
		Configurable: FLAG_TRUE,
		Value:        v,
	}, true)
}

func toPropertyKey(key Value) Value {
	return key.ToString()
}

func (r *Runtime) getVStr(v Value, p unistring.String) Value {
	o := v.ToObject(r)
	return o.self.getStr(p, v)
}

func (r *Runtime) getV(v Value, p Value) Value {
	o := v.ToObject(r)
	return o.get(p, v)
}

type iteratorRecord struct {
	iterator *Object
	next     func(FunctionCall) Value
}

func (r *Runtime) getIterator(obj Value, method func(FunctionCall) Value) *iteratorRecord {
	if method == nil {
		method = toMethod(r.getV(obj, SymIterator))
		if method == nil {
			panic(r.NewTypeError("object is not iterable"))
		}
	}

	iter := r.toObject(method(FunctionCall{
		This: obj,
	}))
	next := toMethod(iter.self.getStr("next", nil))
	return &iteratorRecord{
		iterator: iter,
		next:     next,
	}
}

func (ir *iteratorRecord) iterate(step func(Value)) {
	r := ir.iterator.runtime
	for {
		res := r.toObject(ir.next(FunctionCall{This: ir.iterator}))
		if nilSafe(res.self.getStr("done", nil)).ToBoolean() {
			break
		}
		value := nilSafe(res.self.getStr("value", nil))
		ret := tryFunc(func() {
			step(value)
		})
		if ret != nil {
			_ = tryFunc(func() {
				ir.returnIter()
			})
			panic(ret)
		}
	}
}

func (ir *iteratorRecord) step() (value Value, ex *Exception) {
	r := ir.iterator.runtime
	ex = r.vm.try(func() {
		res := r.toObject(ir.next(FunctionCall{This: ir.iterator}))
		done := nilSafe(res.self.getStr("done", nil)).ToBoolean()
		if !done {
			value = nilSafe(res.self.getStr("value", nil))
		} else {
			ir.close()
		}
	})
	return
}

func (ir *iteratorRecord) returnIter() {
	if ir.iterator == nil {
		return
	}
	retMethod := toMethod(ir.iterator.self.getStr("return", nil))
	if retMethod != nil {
		ir.iterator.runtime.toObject(retMethod(FunctionCall{This: ir.iterator}))
	}
	ir.iterator = nil
	ir.next = nil
}

func (ir *iteratorRecord) close() {
	ir.iterator = nil
	ir.next = nil
}

func (r *Runtime) createIterResultObject(value Value, done bool) Value {
	o := r.NewObject()
	o.self.setOwnStr("value", value, false)
	o.self.setOwnStr("done", r.toBoolean(done), false)
	return o
}

func (r *Runtime) newLazyObject(create func(*Object) objectImpl) *Object {
	val := &Object{runtime: r}
	o := &lazyObject{
		val:    val,
		create: create,
	}
	val.self = o
	return val
}

func (r *Runtime) getHash() *maphash.Hash {
	if r.hash == nil {
		r.hash = &maphash.Hash{}
	}
	return r.hash
}

// leave 当顶层函数正常返回时被调用（即控制被传递到 Runtime 之外）
func (r *Runtime) leave() {
	for {
		jobs := r.jobQueue
		r.jobQueue = nil
		if len(jobs) == 0 {
			break
		}
		for _, job := range jobs {
			job()
		}
	}
}

// leaveAbrupt 当顶层函数返回时被调用（即控制权被传递到Runtime之外），但它是由于一个中断造成的
func (r *Runtime) leaveAbrupt() {
	r.jobQueue = nil
	r.ClearInterrupt()
}

func nilSafe(v Value) Value {
	if v != nil {
		return v
	}
	return _undefined
}

func isArray(object *Object) bool {
	self := object.self
	if proxy, ok := self.(*proxyObject); ok {
		if proxy.target == nil {
			panic(typeError("Cannot perform 'IsArray' on a proxy that has been revoked"))
		}
		return isArray(proxy.target)
	}
	switch self.className() {
	case classArray:
		return true
	default:
		return false
	}
}

func isRegexp(v Value) bool {
	if o, ok := v.(*Object); ok {
		matcher := o.self.getSym(SymMatch, nil)
		if matcher != nil && matcher != _undefined {
			return matcher.ToBoolean()
		}
		_, reg := o.self.(*regexpObject)
		return reg
	}
	return false
}

func limitCallArgs(call FunctionCall, n int) FunctionCall {
	if len(call.Arguments) > n {
		return FunctionCall{This: call.This, Arguments: call.Arguments[:n]}
	} else {
		return call
	}
}

func shrinkCap(newSize, oldCap int) int {
	if oldCap > 8 {
		if cap := oldCap / 2; cap >= newSize {
			return cap
		}
	}
	return oldCap
}

func growCap(newSize, oldSize, oldCap int) int {
	doublecap := oldCap + oldCap
	if newSize > doublecap {
		return newSize
	} else {
		if oldSize < 1024 {
			return doublecap
		} else {
			cap := oldCap
			for 0 < cap && cap < newSize {
				cap += cap / 4
			}
			if cap <= 0 {
				return newSize
			}
			return cap
		}
	}
}

func (r *Runtime) genId() (ret uint64) {
	if r.hash == nil {
		h := r.getHash()
		r.idSeq = h.Sum64()
	}
	if r.idSeq == 0 {
		r.idSeq = 1
	}
	ret = r.idSeq
	r.idSeq++
	return
}

func (r *Runtime) setGlobal(name unistring.String, v Value, strict bool) {
	if ref := r.global.stash.getRefByName(name, strict); ref != nil {
		ref.set(v)
	} else {
		o := r.globalObject.self
		if strict {
			if o.hasOwnPropertyStr(name) {
				o.setOwnStr(name, v, true)
			} else {
				r.throwReferenceError(name)
			}
		} else {
			o.setOwnStr(name, v, false)
		}
	}
}

func (r *Runtime) trackPromiseRejection(p *Promise, operation PromiseRejectionOperation) {
	if r.promiseRejectionTracker != nil {
		r.promiseRejectionTracker(p, operation)
	}
}

func (r *Runtime) callJobCallback(job *jobCallback, this Value, args ...Value) Value {
	return job.callback(FunctionCall{This: this, Arguments: args})
}

func (r *Runtime) invoke(v Value, p unistring.String, args ...Value) Value {
	o := v.ToObject(r)
	return r.toCallable(o.self.getStr(p, nil))(FunctionCall{This: v, Arguments: args})
}

func (r *Runtime) iterableToList(iterable Value, method func(FunctionCall) Value) []Value {
	iter := r.getIterator(iterable, method)
	var values []Value
	iter.iterate(func(item Value) {
		values = append(values, item)
	})
	return values
}

func (r *Runtime) putSpeciesReturnThis(o objectImpl) {
	o._putSym(SymSpecies, &valueProperty{
		getterFunc:   r.newNativeFunc(r.returnThis, nil, "get [Symbol.species]", nil, 0),
		accessor:     true,
		configurable: true,
	})
}

func strToArrayIdx(s unistring.String) uint32 {
	if s == "" {
		return math.MaxUint32
	}
	l := len(s)
	if s[0] == '0' {
		if l == 1 {
			return 0
		}
		return math.MaxUint32
	}
	var n uint32
	if l < 10 {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c < '0' || c > '9' {
				return math.MaxUint32
			}
			n = n*10 + uint32(c-'0')
		}
		return n
	}
	if l > 10 {
		return math.MaxUint32
	}
	c9 := s[9]
	if c9 < '0' || c9 > '9' {
		return math.MaxUint32
	}
	for i := 0; i < 9; i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return math.MaxUint32
		}
		n = n*10 + uint32(c-'0')
	}
	if n >= math.MaxUint32/10+1 {
		return math.MaxUint32
	}
	n *= 10
	n1 := n + uint32(c9-'0')
	if n1 < n {
		return math.MaxUint32
	}

	return n1
}

func strToInt32(s unistring.String) (int32, bool) {
	if s == "" {
		return -1, false
	}
	neg := s[0] == '-'
	if neg {
		s = s[1:]
	}
	l := len(s)
	if s[0] == '0' {
		if l == 1 {
			return 0, !neg
		}
		return -1, false
	}
	var n uint32
	if l < 10 {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c < '0' || c > '9' {
				return -1, false
			}
			n = n*10 + uint32(c-'0')
		}
	} else if l > 10 {
		return -1, false
	} else {
		c9 := s[9]
		if c9 >= '0' {
			if !neg && c9 > '7' || c9 > '8' {
				return -1, false
			}
			for i := 0; i < 9; i++ {
				c := s[i]
				if c < '0' || c > '9' {
					return -1, false
				}
				n = n*10 + uint32(c-'0')
			}
			if n >= math.MaxInt32/10+1 {
				return 0, false
			}
			n = n*10 + uint32(c9-'0')
		} else {
			return -1, false
		}
	}
	if neg {
		return int32(-n), true
	}
	return int32(n), true
}

func strToInt64(s unistring.String) (int64, bool) {
	if s == "" {
		return -1, false
	}
	neg := s[0] == '-'
	if neg {
		s = s[1:]
	}
	l := len(s)
	if s[0] == '0' {
		if l == 1 {
			return 0, !neg
		}
		return -1, false
	}
	var n uint64
	if l < 19 {
		for i := 0; i < len(s); i++ {
			c := s[i]
			if c < '0' || c > '9' {
				return -1, false
			}
			n = n*10 + uint64(c-'0')
		}
	} else if l > 19 {
		return -1, false
	} else {
		c18 := s[18]
		if c18 >= '0' {
			if !neg && c18 > '7' || c18 > '8' {
				return -1, false
			}
			for i := 0; i < 18; i++ {
				c := s[i]
				if c < '0' || c > '9' {
					return -1, false
				}
				n = n*10 + uint64(c-'0')
			}
			if n >= math.MaxInt64/10+1 {
				return 0, false
			}
			n = n*10 + uint64(c18-'0')
		} else {
			return -1, false
		}
	}
	if neg {
		return int64(-n), true
	}
	return int64(n), true
}

func strToInt(s unistring.String) (int, bool) {
	if bits.UintSize == 32 {
		n, ok := strToInt32(s)
		return int(n), ok
	}
	n, ok := strToInt64(s)
	return int(n), ok
}

func strToIntNum(s unistring.String) (int, bool) {
	n, ok := strToInt64(s)
	if n == 0 {
		return 0, ok
	}
	if ok && n >= -maxInt && n <= maxInt {
		if bits.UintSize == 32 {
			if n > math.MaxInt32 || n < math.MinInt32 {
				return 0, false
			}
		}
		return int(n), true
	}
	str := stringValueFromRaw(s)
	if str.ToNumber().toString().SameAs(str) {
		return 0, false
	}
	return -1, false
}

func strToGoIdx(s unistring.String) int {
	if n, ok := strToInt(s); ok {
		return n
	}
	return -1
}

func strToIdx64(s unistring.String) int64 {
	if n, ok := strToInt64(s); ok {
		return n
	}
	return -1
}

func assertCallable(v Value) (func(FunctionCall) Value, bool) {
	if obj, ok := v.(*Object); ok {
		return obj.self.assertCallable()
	}
	return nil, false
}
