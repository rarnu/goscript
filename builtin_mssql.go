package goscript

import (
    "database/sql"
    "fmt"
    _ "github.com/denisenkom/go-mssqldb"
    d0 "github.com/isyscore/isc-gobase/database"
)

type mssqlObject struct {
	baseObject
	db *sql.DB
}

func (r *Runtime) builtinMssql_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*mssqlObject)
	if !ok {
        panic(r.NewTypeError("Method Mssql.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = mo.db.Close()
	return _undefined
}

func (r *Runtime) builtinMssql_exec(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*mssqlObject)
	if !ok {
        panic(r.NewTypeError("Method Mssql.prototype.exec called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	_, err := mo.db.Exec(_sql)
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinMssql_query(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*mssqlObject)
	if !ok {
        panic(r.NewTypeError("Method Mssql.prototype.query called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	rows, err := d0.Query(mo.db, _sql)
	if err != nil {
		return _null
	} else {
		return r.ToValue(rows)
	}
}

func (r *Runtime) builtin_newMssql(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
        panic(r.needNew("Mssql"))
	}
	if len(args) != 5 {
		panic("number of arguments must be 5")
	}

	// 连接数据库
	_host := args[0].toString().String()
	_port := args[1].ToInteger()
	_user := args[2].toString().String()
	_password := args[3].toString().String()
	_database := args[4].toString().String()
    db, err := sql.Open("sqlserver", fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", _user, _password, _host, _port, _database))
	if err != nil {
		return nil
	} else {
        proto := r.getPrototypeFromCtor(newTarget, r.global.Mssql, r.global.MssqlPrototype)
		o := &Object{runtime: r}
		mo := &mssqlObject{
			baseObject: baseObject{
				class:      classMssql,
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

func (r *Runtime) createMssqlProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
    o._putProp("constructor", r.global.Mssql, true, false, true)
    o._putProp("close", r.newNativeFunc(r.builtinMssql_close, nil, "close", nil, 0), true, false, true)
    o._putProp("exec", r.newNativeFunc(r.builtinMssql_exec, nil, "exec", nil, 1), true, false, true)
    o._putProp("query", r.newNativeFunc(r.builtinMssql_query, nil, "query", nil, 1), true, false, true)
    o._putSym(SymToStringTag, valueProp(asciiString(classMssql), false, false, true))
	return o
}

func (r *Runtime) createMssql(val *Object) objectImpl {
    o := r.newNativeConstructOnly(val, r.builtin_newMssql, r.global.MssqlPrototype, "Mssql", 5)
	return o
}

func (r *Runtime) initMssql() {
    r.global.MssqlPrototype = r.newLazyObject(r.createMssqlProto)
    r.global.Mssql = r.newLazyObject(r.createMssql)
    r.addToGlobal("Mssql", r.global.Mssql)
}
