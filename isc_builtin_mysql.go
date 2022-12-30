package goscript

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	d0 "github.com/isyscore/isc-gobase/database"
)

type mysqlObject struct {
	baseObject
	db *sql.DB
}

func (r *Runtime) builtinMysql_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*mysqlObject)
	if !ok {
		panic(r.NewTypeError("Method Mysql.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = mo.db.Close()
	return _undefined
}

func (r *Runtime) builtinMysql_exec(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*mysqlObject)
	if !ok {
		panic(r.NewTypeError("Method Mysql.prototype.exec called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	_, err := mo.db.Exec(_sql)
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinMysql_query(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*mysqlObject)
	if !ok {
		panic(r.NewTypeError("Method Mysql.prototype.query called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_sql := call.Argument(0).toString().String()
	rows, err := d0.Query(mo.db, _sql)
	if err != nil {
		return _null
	} else {
		return r.ToValue(rows)
	}
}

func (r *Runtime) builtin_newMysql(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("Mysql"))
	}
	if len(args) != 5 {
		panic("number of arguments must be 5")
	}

	// 连接数据库
	_host := args[0].toString().String()
	_port := args[1].ToInteger()
	_user := args[2].toString().String()
	_password := args[3].toString().String()
	_dbname := args[4].toString().String()
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True", _user, _password, _host, _port, _dbname))
	if err != nil {
		return nil
	} else {
		proto := r.getPrototypeFromCtor(newTarget, r.global.Mysql, r.global.MysqlPrototype)
		o := &Object{runtime: r}
		mo := &mysqlObject{
			baseObject: baseObject{
				class:      classMysql,
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

func (r *Runtime) createMysqlProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("constructor", r.global.Mysql, true, false, true)
	o._putProp("close", r.newNativeFunc(r.builtinMysql_close, nil, "close", nil, 0), true, false, true)
	o._putProp("exec", r.newNativeFunc(r.builtinMysql_exec, nil, "exec", nil, 1), true, false, true)
	o._putProp("query", r.newNativeFunc(r.builtinMysql_query, nil, "query", nil, 1), true, false, true)
	o._putSym(SymToStringTag, valueProp(asciiString(classMysql), false, false, true))
	return o
}

func (r *Runtime) createMysql(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newMysql, r.global.MysqlPrototype, "Mysql", 5)
	return o
}

func (r *Runtime) initMySQL() {
	r.global.MysqlPrototype = r.newLazyObject(r.createMysqlProto)
	r.global.Mysql = r.newLazyObject(r.createMysql)
	r.addToGlobal("Mysql", r.global.Mysql)
}
