package goscript

import (
	c0 "context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type etcdObject struct {
	baseObject
	cli *clientv3.Client
}

func (r *Runtime) builtinEtcd_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	eo, ok := thisObj.self.(*etcdObject)
	if !ok {
		panic(r.NewTypeError("Method Etcd.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = eo.cli.Close()
	return _undefined
}

func (r *Runtime) builtinEtcd_get(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	eo, ok := thisObj.self.(*etcdObject)
	if !ok {
		panic(r.NewTypeError("Method Etcd.prototype.get called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	gr, err := eo.cli.Get(c0.Background(), call.Argument(0).toString().String(), clientv3.WithPrefix())
	if err != nil {
		return _null
	} else {
		// 返回一个 KV 数组
		var ret0 []map[string]string
		for _, item := range gr.Kvs {
			_m := map[string]string{
				"key":   string(item.Key),
				"value": string(item.Value),
			}
			ret0 = append(ret0, _m)
		}
		return r.ToValue(ret0)
	}
}

func (r *Runtime) builtinEtcd_put(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	eo, ok := thisObj.self.(*etcdObject)
	if !ok {
		panic(r.NewTypeError("Method Etcd.prototype.put called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_val := call.Argument(1).toString().String()
	_, err := eo.cli.Put(c0.Background(), _key, _val)
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(err == nil)
	}
}

func (r *Runtime) builtinEtcd_delete(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	eo, ok := thisObj.self.(*etcdObject)
	if !ok {
		panic(r.NewTypeError("Method Etcd.prototype.delete called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_, err := eo.cli.Delete(c0.Background(), call.Argument(0).toString().String())
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(err == nil)
	}
}

func (r *Runtime) builtinEtcd_sync(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	eo, ok := thisObj.self.(*etcdObject)
	if !ok {
		panic(r.NewTypeError("Method Etcd.prototype.sync called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	err := eo.cli.Sync(c0.Background())
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(err == nil)
	}
}

func (r *Runtime) builtin_newEtcd(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("Etcd"))
	}
	if len(args) != 4 {
		panic("number of arguments must be 4")
	}

	// 连接 etcd
	_host := args[0].toString().String()
	_port := args[1].ToInteger()
	_user := args[2].toString().String()
	_password := args[3].toString().String()
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{fmt.Sprintf("%s:%d", _host, _port)},
		Username:    _user,
		Password:    _password,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil
	} else {
		proto := r.getPrototypeFromCtor(newTarget, r.global.Etcd, r.global.EtcdPrototype)
		o := &Object{runtime: r}
		eo := &etcdObject{
			baseObject: baseObject{
				class:      classEtcd,
				val:        o,
				prototype:  proto,
				extensible: true,
				values:     nil,
			},
			cli: cli,
		}
		o.self = eo
		eo.init()
		return o
	}
}

func (r *Runtime) createEtcdProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("constructor", r.global.Etcd, true, false, true)
	o._putProp("close", r.newNativeFunc(r.builtinEtcd_close, nil, "close", nil, 0), true, false, true)
	o._putProp("get", r.newNativeFunc(r.builtinEtcd_get, nil, "get", nil, 1), true, false, true)
	o._putProp("put", r.newNativeFunc(r.builtinEtcd_put, nil, "put", nil, 2), true, false, true)
	o._putProp("delete", r.newNativeFunc(r.builtinEtcd_delete, nil, "delete", nil, 1), true, false, true)
	o._putProp("sync", r.newNativeFunc(r.builtinEtcd_sync, nil, "sync", nil, 0), true, false, true)
	o._putSym(SymToStringTag, valueProp(asciiString(classEtcd), false, false, true))
	return o
}

func (r *Runtime) createEtcd(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newEtcd, r.global.EtcdPrototype, "Etcd", 4)
	return o
}

func (r *Runtime) initEtcd() {
	r.global.EtcdPrototype = r.newLazyObject(r.createEtcdProto)
	r.global.Etcd = r.newLazyObject(r.createEtcd)
	r.addToGlobal("Etcd", r.global.Etcd)
}
