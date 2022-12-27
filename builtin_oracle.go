package goscript

import (
    "database/sql"
    "fmt"
    d0 "github.com/isyscore/isc-gobase/database"
    _ "github.com/sijms/go-ora/v2"
)

type oracleObject struct {
	baseObject
	db *sql.DB
}

func (r *Runtime) builtinOracle_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*oracleObject)
	if !ok {
        panic(r.NewTypeError("Method Oracle.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = mo.db.Close()
	return _undefined
}

func (r *Runtime) builtinOracle_exec(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*oracleObject)
	if !ok {
        panic(r.NewTypeError("Method Oracle.prototype.exec called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	_, err := mo.db.Exec(_sql)
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinOracle_query(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*oracleObject)
	if !ok {
        panic(r.NewTypeError("Method Oracle.prototype.query called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	rows, err := d0.Query(mo.db, _sql)
	if err != nil {
		return _null
	} else {
		return r.ToValue(rows)
	}
}

func (r *Runtime) builtin_newOracle(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
        panic(r.needNew("Oracle"))
	}
	if len(args) != 5 {
		panic("number of arguments must be 5")
	}

	// 连接数据库
	_host := args[0].toString().String()
	_port := args[1].ToInteger()
	_user := args[2].toString().String()
	_password := args[3].toString().String()
	_service := args[4].toString().String()
	db, err := sql.Open("oracle", fmt.Sprintf("oracle://%s:%s@%s:%d/%s", _user, _password, _host, _port, _service))
	if err != nil {
		return nil
	} else {
		proto := r.getPrototypeFromCtor(newTarget, r.global.Oracle, r.global.OraclePrototype)
		o := &Object{runtime: r}
		mo := &oracleObject{
			baseObject: baseObject{
				class:      classOracle,
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

func (r *Runtime) createOracleProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("constructor", r.global.Oracle, true, false, true)
    o._putProp("close", r.newNativeFunc(r.builtinOracle_close, nil, "close", nil, 0), true, false, true)
    o._putProp("exec", r.newNativeFunc(r.builtinOracle_exec, nil, "exec", nil, 1), true, false, true)
    o._putProp("query", r.newNativeFunc(r.builtinOracle_query, nil, "query", nil, 1), true, false, true)
	o._putSym(SymToStringTag, valueProp(asciiString(classOracle), false, false, true))
	return o
}

func (r *Runtime) createOracle(val *Object) objectImpl {
    o := r.newNativeConstructOnly(val, r.builtin_newOracle, r.global.OraclePrototype, "Oracle", 5)
	return o
}

func (r *Runtime) initOracle() {
    r.global.OraclePrototype = r.newLazyObject(r.createOracleProto)
    r.global.Oracle = r.newLazyObject(r.createOracle)
    r.addToGlobal("Oracle", r.global.Oracle)
}
