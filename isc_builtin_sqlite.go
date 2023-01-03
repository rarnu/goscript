package goscript

import (
	"database/sql"
	d0 "github.com/isyscore/isc-gobase/database"
	_ "github.com/mattn/go-sqlite3"
)

type sqliteObject struct {
	baseObject
	db *sql.DB
}

func (r *Runtime) builtinSQLite_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*sqliteObject)
	if !ok {
		panic(r.NewTypeError("Method SQLite.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = mo.db.Close()
	return _undefined
}

func (r *Runtime) builtinSQLite_exec(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*sqliteObject)
	if !ok {
		panic(r.NewTypeError("Method SQLite.prototype.exec called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	_, err := mo.db.Exec(_sql)
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinSQLite_query(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*sqliteObject)
	if !ok {
		panic(r.NewTypeError("Method SQLite.prototype.query called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	rows, err := d0.Query(mo.db, _sql)
	if err != nil {
		return _null
	} else {
		return r.ToValue(rows)
	}
}

func (r *Runtime) builtin_newSQLite(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("SQLite"))
	}
	if len(args) != 1 {
		panic("number of arguments must be 1")
	}

	// 连接数据库
	// sqlite3 的 host 可以是 ":memory:"，表示该数据库是内存数据库
	_host := args[0].toString().String()
	db, err := sql.Open("sqlite", _host)
	if err != nil {
		return nil
	} else {
		proto := r.getPrototypeFromCtor(newTarget, r.global.SQLite, r.global.SQLitePrototype)
		o := &Object{runtime: r}
		mo := &sqliteObject{
			baseObject: baseObject{
				class:      classSQLite,
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

func (r *Runtime) createSQLiteProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classSQLite)
	o._putProp("constructor", r.global.SQLite, true, false, true)
	o._putProp("close", r.newNativeFunc(r.builtinSQLite_close, nil, "close", nil, 0), true, false, true)
	o._putProp("exec", r.newNativeFunc(r.builtinSQLite_exec, nil, "exec", nil, 1), true, false, true)
	o._putProp("query", r.newNativeFunc(r.builtinSQLite_query, nil, "query", nil, 1), true, false, true)
	o._putSym(SymToStringTag, valueProp(asciiString(classSQLite), false, false, true))
	return o
}

func (r *Runtime) createSQLite(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newSQLite, r.global.SQLitePrototype, "SQLite", 1)
	return o
}

func (r *Runtime) initSQLite() {
	r.global.SQLitePrototype = r.newLazyObject(r.createSQLiteProto)
	r.global.SQLite = r.newLazyObject(r.createSQLite)
	r.addToGlobal("SQLite", r.global.SQLite)
}
