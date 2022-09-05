package goscript

import (
	"fmt"
	"github.com/rarnu/goscript/unistring"
	"reflect"
	"strconv"
)

/*
DynamicObject 是一个接口，代表一个动态对象的处理程序。这样的对象可以使用 Runtime.NewDynamicObject() 方法创建
请注意，Runtime.ToValue() 对 DynamicObject 没有任何特殊处理。创建一个动态对象的唯一方法是使用 Runtime.NewDynamicObject() 方法
这样做是故意的，以避免在这个接口改变时出现不预期且无声的代码中断
*/
type DynamicObject interface {
	Get(key string) Value           // 为 key 获取一个属性值。如果该属性不存在，则返回 nil
	Set(key string, val Value) bool // 为 key 设置一个属性值。如果成功返回 true，否则返回 false
	Has(key string) bool            // 当指定 key 的属性存在时，返回 true
	Delete(key string) bool         // 删除以 key 为名的属性，成功时返回 true（包括不存在的属性）
	Keys() []string                 // 返回一个所有现有 key 的列表。没有重复检查，也没有确定顺序是否符合 https://262.ecma-international.org/#sec-ordinaryownpropertykeys。
}

/*
DynamicArray 是一个接口，代表一个动态数组对象的处理程序。这样的对象可以用 Runtime.NewDynamicArray() 方法创建
任何可以解析为 int 值的整型属性 key 或字符串属性 key（包括负数）都被视为索引，并传递给 DynamicArray 的捕获方法
注意这与普通的 ECMAScript 数组不同，后者只支持 2^32-1 以内的正数索引
动态数组不能是稀疏的，即 hasOwnProperty(num) 对于 num >= 0 && num < Len()，将返回 true
删除这样的属性等同于将其设置为 undefined, 这产生了一个轻微的特殊性，因为 hasOwnProperty() 即使在删除之后，仍然会返回 true
请注意，Runtime.ToValue() 对 DynamicArray 没有任何特殊处理。创建一个动态数组的唯一方法是使用 Runtime.NewDynamicArray()方法
这样做是故意的，以避免在这个接口改变时出现不预期且无声的代码中断
*/
type DynamicArray interface {
	Len() int                    // 返回当前数组的长度
	Get(idx int) Value           // 在索引 idx 处获取一个 item。idx 可以是任何整数，可以是负数，也可以是超过当前长度的
	Set(idx int, val Value) bool // 在索引 idx 处设置一个 item。idx 可以是任何整数，可以是负数，也可以超过当前的长度。当它超出长度时，预期的行为是数组的长度会增加以容纳该 item。数组中的新增部分的所有元素将会被预置为零值
	SetLen(int) bool             // 变更数组的长度，如果长度增加，数组中的新增部分的所有元素将会被预置为零
}

type baseDynamicObject struct {
	val       *Object
	prototype *Object
}

type dynamicObject struct {
	baseDynamicObject
	d DynamicObject
}

type dynamicArray struct {
	baseDynamicObject
	a DynamicArray
}

/*
NewDynamicObject 创建一个由 DynamicObject 处理程序支持的对象
这个对象的特性如下：
1.它的所有属性都是可写、可列举和可配置的数据属性。任何试图定义一个不符合这一点的属性的行为都会失败
2.它总是可扩展的，不能使其成为不可扩展的。尝试执行 Object.preventExtensions() 将会失败
3.它的原型最初被设置为 Object.prototype，但可以使用常规机制（Go 中的 Object.SetPrototype() 或 JS 中的 Object.setPrototypeOf() ）来改变
4.它不能有自己的符号属性，但是它的原型可以。例如，如果你需要一个迭代器支持，你可以创建一个普通的对象，在该对象上设置 Symbol.iterator，然后把它作为一个原型
Export() 返回原始的 DynamicObject

这种机制类似于 ECMAScript Proxy，但是因为所有的属性都是可枚举的，而且对象总是可扩展的，所以不需要进行不变性检查，这就不需要有一个目标对象了，而且效率更高
*/
func (r *Runtime) NewDynamicObject(d DynamicObject) *Object {
	v := &Object{runtime: r}
	o := &dynamicObject{
		d: d,
		baseDynamicObject: baseDynamicObject{
			val:       v,
			prototype: r.global.ObjectPrototype,
		},
	}
	v.self = o
	return v
}

/*
NewSharedDynamicObject 与 Runtime.NewDynamicObject 类似，但是产生的 Object 可以在多个 Runtime 之间共享
该对象的原型为空。提供的 DynamicObject 必须是协程安全的
*/
func NewSharedDynamicObject(d DynamicObject) *Object {
	v := &Object{}
	o := &dynamicObject{
		d: d,
		baseDynamicObject: baseDynamicObject{
			val: v,
		},
	}
	v.self = o
	return v
}

/*
NewDynamicArray 创建一个由 DynamicArray 处理程序支持的数组对象
它与 NewDynamicObject 相似，区别在于：
1.该对象是一个数组（即 Array.isArray() 将返回 true，并且它将有长度属性）
2.对象的原型将被初始设置为 Array.prototype
3.除了长度，该对象不能有任何自己的字符串属性
*/
func (r *Runtime) NewDynamicArray(a DynamicArray) *Object {
	v := &Object{runtime: r}
	o := &dynamicArray{
		a: a,
		baseDynamicObject: baseDynamicObject{
			val:       v,
			prototype: r.global.ArrayPrototype,
		},
	}
	v.self = o
	return v
}

/*
NewSharedDynamicArray 与 Runtime.NewDynamicArray 类似，但产生的 Object 可以在多个 Runtimes 之间共享
该对象的原型为空。如果你需要对它运行 Array 的方法，请使用 Array.prototype.[...].call(a, ...)，提供的 DynamicArray 必须是协程安全的
*/
func NewSharedDynamicArray(a DynamicArray) *Object {
	v := &Object{}
	o := &dynamicArray{
		a: a,
		baseDynamicObject: baseDynamicObject{
			val: v,
		},
	}
	v.self = o
	return v
}

func (*dynamicObject) sortLen() int {
	return 0
}

func (*dynamicObject) sortGet(i int) Value {
	return nil
}

func (*dynamicObject) swap(i int, i2 int) {
}

func (*dynamicObject) className() string {
	return classObject
}

func (o *baseDynamicObject) getParentStr(p unistring.String, receiver Value) Value {
	if proto := o.prototype; proto != nil {
		if receiver == nil {
			return proto.self.getStr(p, o.val)
		}
		return proto.self.getStr(p, receiver)
	}
	return nil
}

func (o *dynamicObject) getStr(p unistring.String, receiver Value) Value {
	prop := o.d.Get(p.String())
	if prop == nil {
		return o.getParentStr(p, receiver)
	}
	return prop
}

func (o *baseDynamicObject) getParentIdx(p valueInt, receiver Value) Value {
	if proto := o.prototype; proto != nil {
		if receiver == nil {
			return proto.self.getIdx(p, o.val)
		}
		return proto.self.getIdx(p, receiver)
	}
	return nil
}

func (o *dynamicObject) getIdx(p valueInt, receiver Value) Value {
	prop := o.d.Get(p.String())
	if prop == nil {
		return o.getParentIdx(p, receiver)
	}
	return prop
}

func (o *baseDynamicObject) getSym(p *Symbol, receiver Value) Value {
	if proto := o.prototype; proto != nil {
		if receiver == nil {
			return proto.self.getSym(p, o.val)
		}
		return proto.self.getSym(p, receiver)
	}
	return nil
}

func (o *dynamicObject) getOwnPropStr(u unistring.String) Value {
	return o.d.Get(u.String())
}

func (o *dynamicObject) getOwnPropIdx(v valueInt) Value {
	return o.d.Get(v.String())
}

func (*baseDynamicObject) getOwnPropSym(*Symbol) Value {
	return nil
}

func (o *dynamicObject) _set(prop string, v Value, throw bool) bool {
	if o.d.Set(prop, v) {
		return true
	}
	typeErrorResult(throw, "'Set' on a dynamic object returned false")
	return false
}

func (o *baseDynamicObject) _setSym(throw bool) {
	typeErrorResult(throw, "Dynamic objects do not support Symbol properties")
}

func (o *dynamicObject) setOwnStr(p unistring.String, v Value, throw bool) bool {
	prop := p.String()
	if !o.d.Has(prop) {
		if proto := o.prototype; proto != nil {
			if res, handled := proto.self.setForeignStr(p, v, o.val, throw); handled {
				return res
			}
		}
	}
	return o._set(prop, v, throw)
}

func (o *dynamicObject) setOwnIdx(p valueInt, v Value, throw bool) bool {
	prop := p.String()
	if !o.d.Has(prop) {
		if proto := o.prototype; proto != nil {
			if res, handled := proto.self.setForeignIdx(p, v, o.val, throw); handled {
				return res
			}
		}
	}
	return o._set(prop, v, throw)
}

func (o *baseDynamicObject) setOwnSym(s *Symbol, v Value, throw bool) bool {
	if proto := o.prototype; proto != nil {
		if res, handled := proto.self.setForeignSym(s, v, o.val, throw); handled {
			return res
		}
	}
	o._setSym(throw)
	return false
}

func (o *baseDynamicObject) setParentForeignStr(p unistring.String, v, receiver Value, throw bool) (res bool, handled bool) {
	if proto := o.prototype; proto != nil {
		if receiver != proto {
			return proto.self.setForeignStr(p, v, receiver, throw)
		}
		return proto.self.setOwnStr(p, v, throw), true
	}
	return false, false
}

func (o *dynamicObject) setForeignStr(p unistring.String, v, receiver Value, throw bool) (res bool, handled bool) {
	prop := p.String()
	if !o.d.Has(prop) {
		return o.setParentForeignStr(p, v, receiver, throw)
	}
	return false, false
}

func (o *baseDynamicObject) setParentForeignIdx(p valueInt, v, receiver Value, throw bool) (res bool, handled bool) {
	if proto := o.prototype; proto != nil {
		if receiver != proto {
			return proto.self.setForeignIdx(p, v, receiver, throw)
		}
		return proto.self.setOwnIdx(p, v, throw), true
	}
	return false, false
}

func (o *dynamicObject) setForeignIdx(p valueInt, v, receiver Value, throw bool) (res bool, handled bool) {
	prop := p.String()
	if !o.d.Has(prop) {
		return o.setParentForeignIdx(p, v, receiver, throw)
	}
	return false, false
}

func (o *baseDynamicObject) setForeignSym(p *Symbol, v, receiver Value, throw bool) (res bool, handled bool) {
	if proto := o.prototype; proto != nil {
		if receiver != proto {
			return proto.self.setForeignSym(p, v, receiver, throw)
		}
		return proto.self.setOwnSym(p, v, throw), true
	}
	return false, false
}

func (o *dynamicObject) hasPropertyStr(u unistring.String) bool {
	if o.hasOwnPropertyStr(u) {
		return true
	}
	if proto := o.prototype; proto != nil {
		return proto.self.hasPropertyStr(u)
	}
	return false
}

func (o *dynamicObject) hasPropertyIdx(idx valueInt) bool {
	if o.hasOwnPropertyIdx(idx) {
		return true
	}
	if proto := o.prototype; proto != nil {
		return proto.self.hasPropertyIdx(idx)
	}
	return false
}

func (o *baseDynamicObject) hasPropertySym(s *Symbol) bool {
	if proto := o.prototype; proto != nil {
		return proto.self.hasPropertySym(s)
	}
	return false
}

func (o *dynamicObject) hasOwnPropertyStr(u unistring.String) bool {
	return o.d.Has(u.String())
}

func (o *dynamicObject) hasOwnPropertyIdx(v valueInt) bool {
	return o.d.Has(v.String())
}

func (*baseDynamicObject) hasOwnPropertySym(_ *Symbol) bool {
	return false
}

func (o *baseDynamicObject) checkDynamicObjectPropertyDescr(name fmt.Stringer, descr PropertyDescriptor, throw bool) bool {
	if descr.Getter != nil || descr.Setter != nil {
		typeErrorResult(throw, "Dynamic objects do not support accessor properties")
		return false
	}
	if descr.Writable == FLAG_FALSE {
		typeErrorResult(throw, "Dynamic object field %q cannot be made read-only", name.String())
		return false
	}
	if descr.Enumerable == FLAG_FALSE {
		typeErrorResult(throw, "Dynamic object field %q cannot be made non-enumerable", name.String())
		return false
	}
	if descr.Configurable == FLAG_FALSE {
		typeErrorResult(throw, "Dynamic object field %q cannot be made non-configurable", name.String())
		return false
	}
	return true
}

func (o *dynamicObject) defineOwnPropertyStr(name unistring.String, desc PropertyDescriptor, throw bool) bool {
	if o.checkDynamicObjectPropertyDescr(name, desc, throw) {
		return o._set(name.String(), desc.Value, throw)
	}
	return false
}

func (o *dynamicObject) defineOwnPropertyIdx(name valueInt, desc PropertyDescriptor, throw bool) bool {
	if o.checkDynamicObjectPropertyDescr(name, desc, throw) {
		return o._set(name.String(), desc.Value, throw)
	}
	return false
}

func (o *baseDynamicObject) defineOwnPropertySym(name *Symbol, desc PropertyDescriptor, throw bool) bool {
	o._setSym(throw)
	return false
}

func (o *dynamicObject) _delete(prop string, throw bool) bool {
	if o.d.Delete(prop) {
		return true
	}
	typeErrorResult(throw, "Could not delete property %q of a dynamic object", prop)
	return false
}

func (o *dynamicObject) deleteStr(name unistring.String, throw bool) bool {
	return o._delete(name.String(), throw)
}

func (o *dynamicObject) deleteIdx(idx valueInt, throw bool) bool {
	return o._delete(idx.String(), throw)
}

func (*baseDynamicObject) deleteSym(_ *Symbol, _ bool) bool {
	return true
}

func (o *baseDynamicObject) toPrimitiveNumber() Value {
	return o.val.genericToPrimitiveNumber()
}

func (o *baseDynamicObject) toPrimitiveString() Value {
	return o.val.genericToPrimitiveString()
}

func (o *baseDynamicObject) toPrimitive() Value {
	return o.val.genericToPrimitive()
}

func (o *baseDynamicObject) assertCallable() (call func(FunctionCall) Value, ok bool) {
	return nil, false
}

func (*baseDynamicObject) assertConstructor() func(args []Value, newTarget *Object) *Object {
	return nil
}

func (o *baseDynamicObject) proto() *Object {
	return o.prototype
}

func (o *baseDynamicObject) setProto(proto *Object, throw bool) bool {
	o.prototype = proto
	return true
}

func (o *baseDynamicObject) hasInstance(v Value) bool {
	panic(newTypeError("Expecting a function in instanceof check, but got a dynamic object"))
}

func (*baseDynamicObject) isExtensible() bool {
	return true
}

func (o *baseDynamicObject) preventExtensions(throw bool) bool {
	typeErrorResult(throw, "Cannot make a dynamic object non-extensible")
	return false
}

type dynamicObjectPropIter struct {
	o         *dynamicObject
	propNames []string
	idx       int
}

func (i *dynamicObjectPropIter) next() (propIterItem, iterNextFunc) {
	for i.idx < len(i.propNames) {
		name := i.propNames[i.idx]
		i.idx++
		if i.o.d.Has(name) {
			return propIterItem{name: newStringValue(name), enumerable: _ENUM_TRUE}, i.next
		}
	}
	return propIterItem{}, nil
}

func (o *dynamicObject) iterateStringKeys() iterNextFunc {
	keys := o.d.Keys()
	return (&dynamicObjectPropIter{
		o:         o,
		propNames: keys,
	}).next
}

func (o *baseDynamicObject) iterateSymbols() iterNextFunc {
	return func() (propIterItem, iterNextFunc) {
		return propIterItem{}, nil
	}
}

func (o *dynamicObject) iterateKeys() iterNextFunc {
	return o.iterateStringKeys()
}

func (o *dynamicObject) export(ctx *objectExportCtx) any {
	return o.d
}

func (o *dynamicObject) exportType() reflect.Type {
	return reflect.TypeOf(o.d)
}

func (o *baseDynamicObject) exportToMap(dst reflect.Value, typ reflect.Type, ctx *objectExportCtx) error {
	return genericExportToMap(o.val, dst, typ, ctx)
}

func (o *baseDynamicObject) exportToArrayOrSlice(dst reflect.Value, typ reflect.Type, ctx *objectExportCtx) error {
	return genericExportToArrayOrSlice(o.val, dst, typ, ctx)
}

func (o *dynamicObject) equal(impl objectImpl) bool {
	if other, ok := impl.(*dynamicObject); ok {
		return o.d == other.d
	}
	return false
}

func (o *dynamicObject) stringKeys(all bool, accum []Value) []Value {
	keys := o.d.Keys()
	if l := len(accum) + len(keys); l > cap(accum) {
		oldAccum := accum
		accum = make([]Value, len(accum), l)
		copy(accum, oldAccum)
	}
	for _, key := range keys {
		accum = append(accum, newStringValue(key))
	}
	return accum
}

func (*baseDynamicObject) symbols(all bool, accum []Value) []Value {
	return accum
}

func (o *dynamicObject) keys(all bool, accum []Value) []Value {
	return o.stringKeys(all, accum)
}

func (*baseDynamicObject) _putProp(name unistring.String, value Value, writable, enumerable, configurable bool) Value {
	return nil
}

func (*baseDynamicObject) _putSym(s *Symbol, prop Value) {
}

func (o *baseDynamicObject) getPrivateEnv(*privateEnvType, bool) *privateElements {
	panic(newTypeError("Dynamic objects cannot have private elements"))
}

func (a *dynamicArray) sortLen() int {
	return a.a.Len()
}

func (a *dynamicArray) sortGet(i int) Value {
	return a.a.Get(i)
}

func (a *dynamicArray) swap(i int, j int) {
	x := a.sortGet(i)
	y := a.sortGet(j)
	a.a.Set(i, y)
	a.a.Set(j, x)
}

func (a *dynamicArray) className() string {
	return classArray
}

func (a *dynamicArray) getStr(p unistring.String, receiver Value) Value {
	if p == "length" {
		return intToValue(int64(a.a.Len()))
	}
	if idx, ok := strToInt(p); ok {
		return a.a.Get(idx)
	}
	return a.getParentStr(p, receiver)
}

func (a *dynamicArray) getIdx(p valueInt, receiver Value) Value {
	if val := a.getOwnPropIdx(p); val != nil {
		return val
	}
	return a.getParentIdx(p, receiver)
}

func (a *dynamicArray) getOwnPropStr(u unistring.String) Value {
	if u == "length" {
		return &valueProperty{
			value:    intToValue(int64(a.a.Len())),
			writable: true,
		}
	}
	if idx, ok := strToInt(u); ok {
		return a.a.Get(idx)
	}
	return nil
}

func (a *dynamicArray) getOwnPropIdx(v valueInt) Value {
	return a.a.Get(toIntStrict(int64(v)))
}

func (a *dynamicArray) _setLen(v Value, throw bool) bool {
	if a.a.SetLen(toIntStrict(v.ToInteger())) {
		return true
	}
	typeErrorResult(throw, "'SetLen' on a dynamic array returned false")
	return false
}

func (a *dynamicArray) setOwnStr(p unistring.String, v Value, throw bool) bool {
	if p == "length" {
		return a._setLen(v, throw)
	}
	if idx, ok := strToInt(p); ok {
		return a._setIdx(idx, v, throw)
	}
	typeErrorResult(throw, "Cannot set property %q on a dynamic array", p.String())
	return false
}

func (a *dynamicArray) _setIdx(idx int, v Value, throw bool) bool {
	if a.a.Set(idx, v) {
		return true
	}
	typeErrorResult(throw, "'Set' on a dynamic array returned false")
	return false
}

func (a *dynamicArray) setOwnIdx(p valueInt, v Value, throw bool) bool {
	return a._setIdx(toIntStrict(int64(p)), v, throw)
}

func (a *dynamicArray) setForeignStr(p unistring.String, v, receiver Value, throw bool) (res bool, handled bool) {
	return a.setParentForeignStr(p, v, receiver, throw)
}

func (a *dynamicArray) setForeignIdx(p valueInt, v, receiver Value, throw bool) (res bool, handled bool) {
	return a.setParentForeignIdx(p, v, receiver, throw)
}

func (a *dynamicArray) hasPropertyStr(u unistring.String) bool {
	if a.hasOwnPropertyStr(u) {
		return true
	}
	if proto := a.prototype; proto != nil {
		return proto.self.hasPropertyStr(u)
	}
	return false
}

func (a *dynamicArray) hasPropertyIdx(idx valueInt) bool {
	if a.hasOwnPropertyIdx(idx) {
		return true
	}
	if proto := a.prototype; proto != nil {
		return proto.self.hasPropertyIdx(idx)
	}
	return false
}

func (a *dynamicArray) _has(idx int) bool {
	return idx >= 0 && idx < a.a.Len()
}

func (a *dynamicArray) hasOwnPropertyStr(u unistring.String) bool {
	if u == "length" {
		return true
	}
	if idx, ok := strToInt(u); ok {
		return a._has(idx)
	}
	return false
}

func (a *dynamicArray) hasOwnPropertyIdx(v valueInt) bool {
	return a._has(toIntStrict(int64(v)))
}

func (a *dynamicArray) defineOwnPropertyStr(name unistring.String, desc PropertyDescriptor, throw bool) bool {
	if a.checkDynamicObjectPropertyDescr(name, desc, throw) {
		if idx, ok := strToInt(name); ok {
			return a._setIdx(idx, desc.Value, throw)
		}
		typeErrorResult(throw, "Cannot define property %q on a dynamic array", name.String())
	}
	return false
}

func (a *dynamicArray) defineOwnPropertyIdx(name valueInt, desc PropertyDescriptor, throw bool) bool {
	if a.checkDynamicObjectPropertyDescr(name, desc, throw) {
		return a._setIdx(toIntStrict(int64(name)), desc.Value, throw)
	}
	return false
}

func (a *dynamicArray) _delete(idx int, throw bool) bool {
	if a._has(idx) {
		a._setIdx(idx, _undefined, throw)
	}
	return true
}

func (a *dynamicArray) deleteStr(name unistring.String, throw bool) bool {
	if idx, ok := strToInt(name); ok {
		return a._delete(idx, throw)
	}
	if a.hasOwnPropertyStr(name) {
		typeErrorResult(throw, "Cannot delete property %q on a dynamic array", name.String())
		return false
	}
	return true
}

func (a *dynamicArray) deleteIdx(idx valueInt, throw bool) bool {
	return a._delete(toIntStrict(int64(idx)), throw)
}

type dynArrayPropIter struct {
	a          DynamicArray
	idx, limit int
}

func (i *dynArrayPropIter) next() (propIterItem, iterNextFunc) {
	if i.idx < i.limit && i.idx < i.a.Len() {
		name := strconv.Itoa(i.idx)
		i.idx++
		return propIterItem{name: asciiString(name), enumerable: _ENUM_TRUE}, i.next
	}

	return propIterItem{}, nil
}

func (a *dynamicArray) iterateStringKeys() iterNextFunc {
	return (&dynArrayPropIter{
		a:     a.a,
		limit: a.a.Len(),
	}).next
}

func (a *dynamicArray) iterateKeys() iterNextFunc {
	return a.iterateStringKeys()
}

func (a *dynamicArray) export(ctx *objectExportCtx) any {
	return a.a
}

func (a *dynamicArray) exportType() reflect.Type {
	return reflect.TypeOf(a.a)
}

func (a *dynamicArray) equal(impl objectImpl) bool {
	if other, ok := impl.(*dynamicArray); ok {
		return a == other
	}
	return false
}

func (a *dynamicArray) stringKeys(all bool, accum []Value) []Value {
	al := a.a.Len()
	l := len(accum) + al
	if all {
		l++
	}
	if l > cap(accum) {
		oldAccum := accum
		accum = make([]Value, len(oldAccum), l)
		copy(accum, oldAccum)
	}
	for i := 0; i < al; i++ {
		accum = append(accum, asciiString(strconv.Itoa(i)))
	}
	if all {
		accum = append(accum, asciiString("length"))
	}
	return accum
}

func (a *dynamicArray) keys(all bool, accum []Value) []Value {
	return a.stringKeys(all, accum)
}
