package goscript

import (
	c0 "context"
	"fmt"
	influxdb "github.com/influxdata/influxdb-client-go/v2"
	iapi "github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"time"
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

type influxdbPointObject struct {
	baseObject
	p *write.Point
}

func (r *Runtime) builtinInfluxDBWrite_writePoint(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbWriteObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBWrite.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_obj := call.Argument(0).baseObject(r)
	if _o0, ok0 := _obj.self.(*influxdbPointObject); ok0 {
		mo.wa.WritePoint(_o0.p)
	}
	return _undefined
}

func (r *Runtime) builtinInfluxDBWrite_writeRecord(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbWriteObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBWrite.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_line := call.Argument(0).toString().String()
	mo.wa.WriteRecord(_line)
	return _undefined
}

func (r *Runtime) builtinInfluxDBWrite_flush(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbWriteObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBWrite.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	mo.wa.Flush()
	return _undefined
}

func (r *Runtime) builtinInfluxDBQuery_query(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbQueryObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBQuery.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_q := call.Argument(0).toString().String()
	var ms []map[string]any
	if ret0, err := mo.qa.Query(c0.Background(), _q); err == nil {
		for ret0.Next() {
			m := map[string]any{
				"measurement": ret0.Record().Measurement(),
				"time":        timeToMsec(ret0.Record().Time()),
				"field":       ret0.Record().Field(),
				"start":       timeToMsec(ret0.Record().Start()),
				"stop":        timeToMsec(ret0.Record().Stop()),
				"value":       ret0.Record().Value(),
				"result":      ret0.Record().Result(),
				"values":      ret0.Record().Values(),
			}
			ms = append(ms, m)
		}
		return r.ToValue(ms)
	} else {
		return _null
	}
}

func (r *Runtime) builtinInfluxDBQuery_queryRaw(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbQueryObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBQuery.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_q := call.Argument(0).toString().String()
	if str, err := mo.qa.QueryRaw(c0.Background(), _q, influxdb.DefaultDialect()); err == nil {
		return r.ToValue(str)
	} else {
		return r.ToValue("")
	}
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

func (r *Runtime) builtinInfluxDBPoint_name(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbPointObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBPoint.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	n := mo.p.Name()
	return r.ToValue(n)
}

func (r *Runtime) builtinInfluxDBPoint_addTag(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbPointObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBPoint.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_n := call.Argument(0).toString().String()
	_a := call.Argument(1).toString().String()
	mo.p.AddTag(_n, _a)
	return mo.val
}

func (r *Runtime) builtinInfluxDBPoint_addField(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbPointObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBPoint.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_n := call.Argument(0).toString().String()
	_a := call.Argument(1).Export()
	mo.p.AddField(_n, _a)
	return mo.val
}

func (r *Runtime) builtinInfluxDBPoint_time(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbPointObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBPoint.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	t := mo.p.Time()
	return intToValue(timeToMsec(t))
}

func (r *Runtime) builtinInfluxDBPoint_setTime(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbPointObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBPoint.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	t := call.Argument(0).ToNumber()
	if !IsNaN(t) {
		mo.p.SetTime(time.UnixMilli(t.ToInteger()))
	}
	return mo.val
}

func (r *Runtime) builtinInfluxDBPoint_fieldList(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbPointObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBPoint.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_m := mo.p.FieldList()
	m := map[string]any{}
	for _, item := range _m {
		m[item.Key] = item.Value
	}
	return r.ToValue(m)
}

func (r *Runtime) builtinInfluxDBPoint_tagList(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	mo, ok := thisObj.self.(*influxdbPointObject)
	if !ok {
		panic(r.NewTypeError("Method InfluxDBPoint.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_m := mo.p.TagList()
	m := map[string]any{}
	for _, item := range _m {
		m[item.Key] = item.Value
	}
	return r.ToValue(m)
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

func (r *Runtime) builtin_newInfluxDBPoint(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("InfluxDBPoint"))
	}
	_me := args[0].toString().String()
	_tags := args[1].Export()
	_points := args[2].Export()
	ptags := map[string]string{}
	ppoints := _points.(map[string]any)
	for k, v := range _tags.(map[string]any) {
		ptags[k] = fmt.Sprintf("%v", v)
	}
	pt := influxdb.NewPoint(_me, ptags, ppoints, time.Now())
	proto := r.getPrototypeFromCtor(newTarget, r.global.InfluxDBPoint, r.global.InfluxDBPointPrototype)
	o := &Object{runtime: r}
	mo := &influxdbPointObject{
		baseObject: baseObject{
			class:      classInfluxDBPoint,
			val:        o,
			prototype:  proto,
			extensible: true,
			values:     nil,
		},
		p: pt,
	}
	o.self = mo
	mo.init()
	return o
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

func (r *Runtime) createInfluxDBPointProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("constructor", r.global.InfluxDBPoint, true, false, true)
	o._putProp("name", r.newNativeFunc(r.builtinInfluxDBPoint_name, nil, "name", nil, 0), true, false, true)
	o._putProp("addTag", r.newNativeFunc(r.builtinInfluxDBPoint_addTag, nil, "addTag", nil, 2), true, false, true)
	o._putProp("addField", r.newNativeFunc(r.builtinInfluxDBPoint_addField, nil, "addField", nil, 2), true, false, true)
	o._putProp("time", r.newNativeFunc(r.builtinInfluxDBPoint_time, nil, "time", nil, 0), true, false, true)
	o._putProp("setTime", r.newNativeFunc(r.builtinInfluxDBPoint_setTime, nil, "setTime", nil, 1), true, false, true)
	o._putProp("fieldList", r.newNativeFunc(r.builtinInfluxDBPoint_fieldList, nil, "fieldList", nil, 0), true, false, true)
	o._putProp("tagList", r.newNativeFunc(r.builtinInfluxDBPoint_tagList, nil, "tagList", nil, 0), true, false, true)
	return o
}

func (r *Runtime) createInfluxDBWriteProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("writeRecord", r.newNativeFunc(r.builtinInfluxDBWrite_writeRecord, nil, "writeRecord", nil, 1), true, false, true)
	o._putProp("writePoint", r.newNativeFunc(r.builtinInfluxDBWrite_writePoint, nil, "writePoint", nil, 1), true, false, true)
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

func (r *Runtime) createInfluxDBPoint(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newInfluxDBPoint, r.global.InfluxDBPointPrototype, "InfluxDBPoint", 3)
	return o
}

func (r *Runtime) initInfluxDB() {
	r.global.InfluxDBPrototype = r.newLazyObject(r.createInfluxDBProto)
	r.global.InfluxDB = r.newLazyObject(r.createInfluxDB)
	r.addToGlobal("InfluxDB", r.global.InfluxDB)
	// 为子对象注册 prototype，需要注意的是，这两个子对象没有构造函数，无法用 new 来实例化
	r.global.InfluxDBWritePrototype = r.newLazyObject(r.createInfluxDBWriteProto)
	r.global.InfluxDBQueryPrototype = r.newLazyObject(r.createInfluxDBQueryProto)
	// Point
	r.global.InfluxDBPointPrototype = r.newLazyObject(r.createInfluxDBPointProto)
	r.global.InfluxDBPoint = r.newLazyObject(r.createInfluxDBPoint)
	r.addToGlobal("InfluxDBPoint", r.global.InfluxDBPoint)
}
