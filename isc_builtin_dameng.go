package goscript

import (
    "database/sql"
    "fmt"
    _ "gitee.com/chunanyong/dm"
    d0 "github.com/isyscore/isc-gobase/database"
)

type damengObject struct {
	baseObject
	db *sql.DB
}

func (r *Runtime) builtinDameng_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*damengObject)
	if !ok {
        panic(r.NewTypeError("Method Dameng.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = mo.db.Close()
	return _undefined
}

func (r *Runtime) builtinDameng_exec(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*damengObject)
	if !ok {
        panic(r.NewTypeError("Method Dameng.prototype.exec called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	_, err := mo.db.Exec(_sql)
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinDameng_query(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*damengObject)
	if !ok {
        panic(r.NewTypeError("Method Dameng.prototype.query called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	rows, err := d0.Query(mo.db, _sql)
	if err != nil {
		return _null
	} else {
		return r.ToValue(rows)
	}
}

func (r *Runtime) builtin_newDameng(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
        panic(r.needNew("Dameng"))
	}
	if len(args) != 4 {
		panic("number of arguments must be 4")
	}

	// 连接数据库
	_host := args[0].toString().String()
	_port := args[1].ToInteger()
	_user := args[2].toString().String()
	_password := args[3].toString().String()
    db, err := sql.Open("dm", fmt.Sprintf("dm://%s:%s@%s:%d", _user, _password, _host, _port))
	if err != nil {
		return nil
	} else {
        proto := r.getPrototypeFromCtor(newTarget, r.global.Dameng, r.global.DamengPrototype)
		o := &Object{runtime: r}
		mo := &damengObject{
			baseObject: baseObject{
                class:      classDameng,
				val:        o,
				prototype:  proto,
				extensible: true,
				values:     nil,
			},
			db: db,
		}
		o.self = mo
		mo.init()
		return o
	}
}

func (r *Runtime) createDamengProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
    o._putProp("constructor", r.global.Dameng, true, false, true)
    o._putProp("close", r.newNativeFunc(r.builtinDameng_close, nil, "close", nil, 0), true, false, true)
    o._putProp("exec", r.newNativeFunc(r.builtinDameng_exec, nil, "exec", nil, 1), true, false, true)
    o._putProp("query", r.newNativeFunc(r.builtinDameng_query, nil, "query", nil, 1), true, false, true)
    o._putSym(SymToStringTag, valueProp(asciiString(classDameng), false, false, true))
	return o
}

func (r *Runtime) createDameng(val *Object) objectImpl {
    o := r.newNativeConstructOnly(val, r.builtin_newDameng, r.global.DamengPrototype, "Dameng", 4)
	return o
}

func (r *Runtime) initDameng() {
    r.global.DamengPrototype = r.newLazyObject(r.createDamengProto)
    r.global.Dameng = r.newLazyObject(r.createDameng)
    r.addToGlobal("Dameng", r.global.Dameng)
}
