package goscript

import (
	"github.com/rarnu/goscript/unistring"
)

type nativeProxyHandler struct {
	handler *ProxyTrapConfig
}

func (h *nativeProxyHandler) getPrototypeOf(target *Object) (Value, bool) {
	if trap := h.handler.GetPrototypeOf; trap != nil {
		return trap(target), true
	}
	return nil, false
}

func (h *nativeProxyHandler) setPrototypeOf(target *Object, proto *Object) (bool, bool) {
	if trap := h.handler.SetPrototypeOf; trap != nil {
		return trap(target, proto), true
	}
	return false, false
}

func (h *nativeProxyHandler) isExtensible(target *Object) (bool, bool) {
	if trap := h.handler.IsExtensible; trap != nil {
		return trap(target), true
	}
	return false, false
}

func (h *nativeProxyHandler) preventExtensions(target *Object) (bool, bool) {
	if trap := h.handler.PreventExtensions; trap != nil {
		return trap(target), true
	}
	return false, false
}

func (h *nativeProxyHandler) getOwnPropertyDescriptorStr(target *Object, prop unistring.String) (Value, bool) {
	if trap := h.handler.GetOwnPropertyDescriptorIdx; trap != nil {
		if idx, ok := strToInt(prop); ok {
			desc := trap(target, idx)
			return desc.toValue(target.runtime), true
		}
	}
	if trap := h.handler.GetOwnPropertyDescriptor; trap != nil {
		desc := trap(target, prop.String())
		return desc.toValue(target.runtime), true
	}
	return nil, false
}

func (h *nativeProxyHandler) getOwnPropertyDescriptorIdx(target *Object, prop valueInt) (Value, bool) {
	if trap := h.handler.GetOwnPropertyDescriptorIdx; trap != nil {
		desc := trap(target, toIntStrict(int64(prop)))
		return desc.toValue(target.runtime), true
	}
	if trap := h.handler.GetOwnPropertyDescriptor; trap != nil {
		desc := trap(target, prop.String())
		return desc.toValue(target.runtime), true
	}
	return nil, false
}

func (h *nativeProxyHandler) getOwnPropertyDescriptorSym(target *Object, prop *Symbol) (Value, bool) {
	if trap := h.handler.GetOwnPropertyDescriptorSym; trap != nil {
		desc := trap(target, prop)
		return desc.toValue(target.runtime), true
	}
	return nil, false
}

func (h *nativeProxyHandler) definePropertyStr(target *Object, prop unistring.String, desc PropertyDescriptor) (bool, bool) {
	if trap := h.handler.DefinePropertyIdx; trap != nil {
		if idx, ok := strToInt(prop); ok {
			return trap(target, idx, desc), true
		}
	}
	if trap := h.handler.DefineProperty; trap != nil {
		return trap(target, prop.String(), desc), true
	}
	return false, false
}

func (h *nativeProxyHandler) definePropertyIdx(target *Object, prop valueInt, desc PropertyDescriptor) (bool, bool) {
	if trap := h.handler.DefinePropertyIdx; trap != nil {
		return trap(target, toIntStrict(int64(prop)), desc), true
	}
	if trap := h.handler.DefineProperty; trap != nil {
		return trap(target, prop.String(), desc), true
	}
	return false, false
}

func (h *nativeProxyHandler) definePropertySym(target *Object, prop *Symbol, desc PropertyDescriptor) (bool, bool) {
	if trap := h.handler.DefinePropertySym; trap != nil {
		return trap(target, prop, desc), true
	}
	return false, false
}

func (h *nativeProxyHandler) hasStr(target *Object, prop unistring.String) (bool, bool) {
	if trap := h.handler.HasIdx; trap != nil {
		if idx, ok := strToInt(prop); ok {
			return trap(target, idx), true
		}
	}
	if trap := h.handler.Has; trap != nil {
		return trap(target, prop.String()), true
	}
	return false, false
}

func (h *nativeProxyHandler) hasIdx(target *Object, prop valueInt) (bool, bool) {
	if trap := h.handler.HasIdx; trap != nil {
		return trap(target, toIntStrict(int64(prop))), true
	}
	if trap := h.handler.Has; trap != nil {
		return trap(target, prop.String()), true
	}
	return false, false
}

func (h *nativeProxyHandler) hasSym(target *Object, prop *Symbol) (bool, bool) {
	if trap := h.handler.HasSym; trap != nil {
		return trap(target, prop), true
	}
	return false, false
}

func (h *nativeProxyHandler) getStr(target *Object, prop unistring.String, receiver Value) (Value, bool) {
	if trap := h.handler.GetIdx; trap != nil {
		if idx, ok := strToInt(prop); ok {
			return trap(target, idx, receiver), true
		}
	}
	if trap := h.handler.Get; trap != nil {
		return trap(target, prop.String(), receiver), true
	}
	return nil, false
}

func (h *nativeProxyHandler) getIdx(target *Object, prop valueInt, receiver Value) (Value, bool) {
	if trap := h.handler.GetIdx; trap != nil {
		return trap(target, toIntStrict(int64(prop)), receiver), true
	}
	if trap := h.handler.Get; trap != nil {
		return trap(target, prop.String(), receiver), true
	}
	return nil, false
}

func (h *nativeProxyHandler) getSym(target *Object, prop *Symbol, receiver Value) (Value, bool) {
	if trap := h.handler.GetSym; trap != nil {
		return trap(target, prop, receiver), true
	}
	return nil, false
}

func (h *nativeProxyHandler) setStr(target *Object, prop unistring.String, value Value, receiver Value) (bool, bool) {
	if trap := h.handler.SetIdx; trap != nil {
		if idx, ok := strToInt(prop); ok {
			return trap(target, idx, value, receiver), true
		}
	}
	if trap := h.handler.Set; trap != nil {
		return trap(target, prop.String(), value, receiver), true
	}
	return false, false
}

func (h *nativeProxyHandler) setIdx(target *Object, prop valueInt, value Value, receiver Value) (bool, bool) {
	if trap := h.handler.SetIdx; trap != nil {
		return trap(target, toIntStrict(int64(prop)), value, receiver), true
	}
	if trap := h.handler.Set; trap != nil {
		return trap(target, prop.String(), value, receiver), true
	}
	return false, false
}

func (h *nativeProxyHandler) setSym(target *Object, prop *Symbol, value Value, receiver Value) (bool, bool) {
	if trap := h.handler.SetSym; trap != nil {
		return trap(target, prop, value, receiver), true
	}
	return false, false
}

func (h *nativeProxyHandler) deleteStr(target *Object, prop unistring.String) (bool, bool) {
	if trap := h.handler.DeletePropertyIdx; trap != nil {
		if idx, ok := strToInt(prop); ok {
			return trap(target, idx), true
		}
	}
	if trap := h.handler.DeleteProperty; trap != nil {
		return trap(target, prop.String()), true
	}
	return false, false
}

func (h *nativeProxyHandler) deleteIdx(target *Object, prop valueInt) (bool, bool) {
	if trap := h.handler.DeletePropertyIdx; trap != nil {
		return trap(target, toIntStrict(int64(prop))), true
	}
	if trap := h.handler.DeleteProperty; trap != nil {
		return trap(target, prop.String()), true
	}
	return false, false
}

func (h *nativeProxyHandler) deleteSym(target *Object, prop *Symbol) (bool, bool) {
	if trap := h.handler.DeletePropertySym; trap != nil {
		return trap(target, prop), true
	}
	return false, false
}

func (h *nativeProxyHandler) ownKeys(target *Object) (*Object, bool) {
	if trap := h.handler.OwnKeys; trap != nil {
		return trap(target), true
	}
	return nil, false
}

func (h *nativeProxyHandler) apply(target *Object, this Value, args []Value) (Value, bool) {
	if trap := h.handler.Apply; trap != nil {
		return trap(target, this, args), true
	}
	return nil, false
}

func (h *nativeProxyHandler) construct(target *Object, args []Value, newTarget *Object) (Value, bool) {
	if trap := h.handler.Construct; trap != nil {
		return trap(target, args, newTarget), true
	}
	return nil, false
}

func (h *nativeProxyHandler) toObject(runtime *Runtime) *Object {
	return runtime.ToValue(h.handler).ToObject(runtime)
}

func (r *Runtime) newNativeProxyHandler(nativeHandler *ProxyTrapConfig) proxyHandler {
	return &nativeProxyHandler{handler: nativeHandler}
}

// ProxyTrapConfig 为实现 Proxy 捕获提供了一个简化友好的 Go API
// 如果定义了 *Idx 捕获，它就会被调用来处理整数属性键，包括负数。注意，这只包括代表典型整数的字符串属性键（即 "0"、"123"，但不包括 "00"、"01"、"1" 或 "-0"）
// 为了提高效率，超过 2^53 的整数的字符串不会被检查是否是典型的，即 *Idx 捕获将收到 "9007199254740993" 和 "9007199254740994"
// 尽管前者在 ECMAScript 中不是典型的表示（Number("9007199254740993") === 9007199254740992）
// 参考 https://262.ecma-international.org/#sec-canonicalnumericindexstring
// 如果没有设置 *Idx 捕获，就会使用相应的字符串捕获
type ProxyTrapConfig struct {
	GetPrototypeOf              func(target *Object) (prototype *Object)                                                // 针对 Object.getPrototypeOf, Reflect.getPrototypeOf, __proto__, Object.prototype.isPrototypeOf, instanceof 的捕获
	SetPrototypeOf              func(target *Object, prototype *Object) (success bool)                                  // 针对 Object.setPrototypeOf, Reflect.setPrototypeOf 的捕获
	IsExtensible                func(target *Object) (success bool)                                                     // 针对 Object.isExtensible, Reflect.isExtensible 的捕获
	PreventExtensions           func(target *Object) (success bool)                                                     // 针对 Object.preventExtensions, Reflect.preventExtensions 的捕获
	GetOwnPropertyDescriptor    func(target *Object, prop string) (propertyDescriptor PropertyDescriptor)               // 针对 Object.getOwnPropertyDescriptor, Reflect.getOwnPropertyDescriptor (字符串属性) 的捕获
	GetOwnPropertyDescriptorIdx func(target *Object, prop int) (propertyDescriptor PropertyDescriptor)                  // 针对 Object.getOwnPropertyDescriptor, Reflect.getOwnPropertyDescriptor (整型属性) 的捕获
	GetOwnPropertyDescriptorSym func(target *Object, prop *Symbol) (propertyDescriptor PropertyDescriptor)              // 针对 Object.getOwnPropertyDescriptor, Reflect.getOwnPropertyDescriptor (符号属性) 的捕获
	DefineProperty              func(target *Object, key string, propertyDescriptor PropertyDescriptor) (success bool)  // 针对 Object.defineProperty, Reflect.defineProperty (字符串属性) 的捕获
	DefinePropertyIdx           func(target *Object, key int, propertyDescriptor PropertyDescriptor) (success bool)     // 针对 Object.defineProperty, Reflect.defineProperty (整型属性) 的捕获
	DefinePropertySym           func(target *Object, key *Symbol, propertyDescriptor PropertyDescriptor) (success bool) // 针对 Object.defineProperty, Reflect.defineProperty (符号属性) 的捕获
	Has                         func(target *Object, property string) (available bool)                                  // 针对 in，with 操作符，Reflect.has (字符串属性) 的捕获
	HasIdx                      func(target *Object, property int) (available bool)                                     // 针对 in，with 操作符，Reflect.has (整型属性) 的捕获
	HasSym                      func(target *Object, property *Symbol) (available bool)                                 // 针对 in，with 操作符，Reflect.has (符号属性) 的捕获
	Get                         func(target *Object, property string, receiver Value) (value Value)                     // 针对读取属性值，Reflect.get (字符串属性) 的捕获
	GetIdx                      func(target *Object, property int, receiver Value) (value Value)                        // 针对读取属性值，Reflect.get (整型属性) 的捕获
	GetSym                      func(target *Object, property *Symbol, receiver Value) (value Value)                    // 针对读取属性值，Reflect.get (符号属性) 的捕获
	Set                         func(target *Object, property string, value Value, receiver Value) (success bool)       // 针对设置属性值，Reflect.set (字符串属性) 的捕获
	SetIdx                      func(target *Object, property int, value Value, receiver Value) (success bool)          // 针对设置属性值，Reflect.set (整型属性) 的捕获
	SetSym                      func(target *Object, property *Symbol, value Value, receiver Value) (success bool)      // 针对设置属性值，Reflect.set (符号属性) 的捕获
	DeleteProperty              func(target *Object, property string) (success bool)                                    // 针对 delete 操作符，Reflect.deleteProperty (字符串属性) 的捕获
	DeletePropertyIdx           func(target *Object, property int) (success bool)                                       // 针对 delete 操作符，Reflect.deleteProperty (整型属性) 的捕获
	DeletePropertySym           func(target *Object, property *Symbol) (success bool)                                   // 针对 delete 操作符，Reflect.deleteProperty (符号属性) 的捕获
	OwnKeys                     func(target *Object) (object *Object)                                                   // 针对 Object.getOwnPropertyNames, Object.getOwnPropertySymbols, Object.keys, Reflect.ownKeys 的捕获
	Apply                       func(target *Object, this Value, argumentsList []Value) (value Value)                   // 针对函数调用，Function.prototype.apply, Function.prototype.call, Reflect.apply 的捕获
	Construct                   func(target *Object, argumentsList []Value, newTarget *Object) (value *Object)          // 针对 new 操作符，Reflect.construct 的捕获
}

func (r *Runtime) newProxy(args []Value, proto *Object) *Object {
	if len(args) >= 2 {
		if target, ok := args[0].(*Object); ok {
			if proxyHandler, ok := args[1].(*Object); ok {
				return r.newProxyObject(target, proxyHandler, proto).val
			}
		}
	}
	panic(r.NewTypeError("Cannot create proxy with a non-object as target or handler"))
}

func (r *Runtime) builtin_newProxy(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("Proxy"))
	}
	return r.newProxy(args, r.getPrototypeFromCtor(newTarget, r.global.Proxy, r.global.ObjectPrototype))
}

func (r *Runtime) NewProxy(target *Object, nativeHandler *ProxyTrapConfig) Proxy {
	if p, ok := target.self.(*proxyObject); ok {
		if p.handler == nil {
			panic(r.NewTypeError("Cannot create proxy with a revoked proxy as target"))
		}
	}
	handler := r.newNativeProxyHandler(nativeHandler)
	proxy := r._newProxyObject(target, handler, nil)
	return Proxy{proxy: proxy}
}

func (r *Runtime) builtin_proxy_revocable(call FunctionCall) Value {
	if len(call.Arguments) >= 2 {
		if target, ok := call.Argument(0).(*Object); ok {
			if proxyHandler, ok := call.Argument(1).(*Object); ok {
				proxy := r.newProxyObject(target, proxyHandler, nil)
				revoke := r.newNativeFunc(func(FunctionCall) Value {
					proxy.revoke()
					return _undefined
				}, nil, "", nil, 0)
				ret := r.NewObject()
				ret.self._putProp("proxy", proxy.val, true, true, true)
				ret.self._putProp("revoke", revoke, true, true, true)
				return ret
			}
		}
	}
	panic(r.NewTypeError("Cannot create proxy with a non-object as target or handler"))
}

func (r *Runtime) createProxy(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newProxy, nil, "Proxy", 2)

	o._putProp("revocable", r.newNativeFunc(r.builtin_proxy_revocable, nil, "revocable", nil, 2), true, false, true)
	return o
}

func (r *Runtime) initProxy() {
	r.global.Proxy = r.newLazyObject(r.createProxy)
	r.addToGlobal("Proxy", r.global.Proxy)
}
