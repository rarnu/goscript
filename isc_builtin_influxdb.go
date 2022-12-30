package goscript

import (
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	iapi "github.com/influxdata/influxdb-client-go/v2/api"
)

type influxdbObject struct {
	baseObject
	cli influxdb.Client
}

type influxdbWriteObject struct {
	baseObject
	wa iapi.WriteAPI
}

type influxdbQueryObject struct {
	baseObject
	qa iapi.QueryAPI
}

func (r *Runtime) builtinInfluxDBWrite_writeRecord(call FunctionCall) Value {
	return _undefined
}

func (r *Runtime) builtinInfluxDBWrite_flush(call FunctionCall) Value {
	return _undefined
}

func (r *Runtime) builtinInfluxDBQuery_query(call FunctionCall) Value {
	return _undefined
}

func (r *Runtime) builtinInfluxDBQuery_queryRaw(call FunctionCall) Value {
	return _undefined
}

func (r *Runtime) builtinInfluxDB_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDB.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	mo.cli.Close()
	return _undefined
}

// buildInfluxDBSubComp 把 InfluxDB 的子对象包装出去
// js 需要有 prototype 才识别对象类型，并可以正确调用到子对象的方法
func (r *Runtime) buildInfluxDBSubComp(i any, isQuery bool) Value {
	o := &Object{runtime: r}
	if isQuery {
		obj := &influxdbQueryObject{
			baseObject: baseObject{class: classInfluxDBQuery, val: o, prototype: r.global.InfluxDBQueryPrototype, extensible: true, values: nil},
			qa:         i.(iapi.QueryAPI),
		}
		o.self = obj
		obj.init()
	} else {
		obj := &influxdbWriteObject{
			baseObject: baseObject{class: classInfluxDBWrite, val: o, prototype: r.global.InfluxDBWritePrototype, extensible: true, values: nil},
			wa:         i.(iapi.WriteAPI),
		}
		o.self = obj
		obj.init()
	}
	return r.ToValue(o)
}

func (r *Runtime) builtinInfluxDB_write(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDB.prototype.exec called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_org := call.Argument(0).toString().String()
	_buc := call.Argument(1).toString().String()
	wa := mo.cli.WriteAPI(_org, _buc)
	return r.buildInfluxDBSubComp(wa, false)
}

func (r *Runtime) builtinInfluxDB_query(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDB.prototype.query called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_org := call.Argument(0).toString().String()
	qa := mo.cli.QueryAPI(_org)
	return r.buildInfluxDBSubComp(qa, true)
}

func (r *Runtime) builtin_newInfluxDB(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("InfluxDB"))
	}
	if len(args) != 2 {
		panic("number of arguments must be 2")
	}

	// 连接数据库
	_url := args[0].toString().String()
	_token := args[1].toString().String()
	cli := influxdb.NewClient(_url, _token)
	if cli == nil {
		return nil
	} else {
		proto := r.getPrototypeFromCtor(newTarget, r.global.InfluxDB, r.global.InfluxDBPrototype)
		o := &Object{runtime: r}
		mo := &influxdbObject{
			baseObject: baseObject{
				class:      classInfluxDB,
				val:        o,
				prototype:  proto,
				extensible: true,
				values:     nil,
			},
			cli: cli,
		}
		o.self = mo
		mo.init()
		return o
	}
}

func (r *Runtime) createInfluxDBProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("constructor", r.global.InfluxDB, true, false, true)
	o._putProp("close", r.newNativeFunc(r.builtinInfluxDB_close, nil, "close", nil, 0), true, false, true)
	o._putProp("write", r.newNativeFunc(r.builtinInfluxDB_write, nil, "write", nil, 2), true, false, true)
	o._putProp("query", r.newNativeFunc(r.builtinInfluxDB_query, nil, "query", nil, 1), true, false, true)
	o._putSym(SymToStringTag, valueProp(asciiString(classInfluxDB), false, false, true))
	return o
}

func (r *Runtime) createInfluxDBWriteProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("writeRecord", r.newNativeFunc(r.builtinInfluxDBWrite_writeRecord, nil, "writeRecord", nil, 1), true, false, true)
	o._putProp("flush", r.newNativeFunc(r.builtinInfluxDBWrite_flush, nil, "flush", nil, 0), true, false, true)
	o._putSym(SymToStringTag, valueProp(asciiString(classInfluxDBWrite), false, false, true))
	return o
}

func (r *Runtime) createInfluxDBQueryProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("query", r.newNativeFunc(r.builtinInfluxDBQuery_query, nil, "query", nil, 1), true, false, true)
	o._putProp("queryRaw", r.newNativeFunc(r.builtinInfluxDBQuery_queryRaw, nil, "queryRaw", nil, 1), true, false, true)
	o._putSym(SymToStringTag, valueProp(asciiString(classInfluxDBQuery), false, false, true))
	return o
}

func (r *Runtime) createInfluxDB(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newInfluxDB, r.global.InfluxDBPrototype, "InfluxDB", 2)
	return o
}

func (r *Runtime) initInfluxDB() {
	r.global.InfluxDBPrototype = r.newLazyObject(r.createInfluxDBProto)
	r.global.InfluxDB = r.newLazyObject(r.createInfluxDB)
	r.addToGlobal("InfluxDB", r.global.InfluxDB)
	// 为子对象注册 prototype，需要注意的是，这两个子对象没有构造函数，无法用 new 来实例化
	r.global.InfluxDBWritePrototype = r.newLazyObject(r.createInfluxDBWriteProto)
	r.global.InfluxDBQueryPrototype = r.newLazyObject(r.createInfluxDBQueryProto)
}
