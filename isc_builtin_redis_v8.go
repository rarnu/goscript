package goscript

import (
	c0 "context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

type redisV8Object struct {
	baseObject
	cli *redis.Client
}

func (r *Runtime) builtinRedisV8_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = ro.cli.Close()
	return _undefined
}

func (r *Runtime) builtinRedisV8_get(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.get called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.Get(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedisV8_set(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.set called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_val := call.Argument(1).Export()
	_expire := call.Argument(2).ToInteger()
	err := ro.cli.Set(c0.Background(), _key, _val, time.Millisecond*time.Duration(_expire)).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedisV8_del(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.del called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	i, err := ro.cli.Del(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_ttl(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.ttl called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	d, err := ro.cli.TTL(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(int64(d))
	}
}

func (r *Runtime) builtinRedisV8_llen(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.llen called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	l, err := ro.cli.LLen(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	}
	return intToValue(l)
}

func (r *Runtime) builtinRedisV8_lrange(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.lrange called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_start := call.Argument(1).ToInteger()
	_stop := call.Argument(2).ToInteger()
	s, err := ro.cli.LRange(c0.Background(), _key, _start, _stop).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedisV8_lindex(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.lindex called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_idx := call.Argument(1).ToInteger()
	s, err := ro.cli.LIndex(c0.Background(), _key, _idx).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedisV8_lpush(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.lpush called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_obj := call.Argument(1).Export()
	i, err := ro.cli.LPush(c0.Background(), _key, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_lpop(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.lpop called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.LPop(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedisV8_lset(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.lset called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_idx := call.Argument(1).ToInteger()
	_obj := call.Argument(2).Export()
	err := ro.cli.LSet(c0.Background(), _key, _idx, _obj).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedisV8_ltrim(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.ltrim called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_start := call.Argument(1).ToInteger()
	_stop := call.Argument(2).ToInteger()
	err := ro.cli.LTrim(c0.Background(), _key, _start, _stop).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedisV8_lrem(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.lrem called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_idx := call.Argument(1).ToInteger()
	_obj := call.Argument(2).Export()
	i, err := ro.cli.LRem(c0.Background(), _key, _idx, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_linsertAfter(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.linsertAfter called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_pivot := call.Argument(1).Export()
	_val := call.Argument(2).Export()
	i, err := ro.cli.LInsertAfter(c0.Background(), _key, _pivot, _val).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_linsertBefore(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.linsertBefore called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_pivot := call.Argument(1).Export()
	_val := call.Argument(2).Export()
	i, err := ro.cli.LInsertBefore(c0.Background(), _key, _pivot, _val).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_sadd(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.sadd called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_obj := call.Argument(1).Export()
	i, err := ro.cli.SAdd(c0.Background(), _key, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_scard(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.scard called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	i, err := ro.cli.SCard(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_sdiff(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.sdiff called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key1 := call.Argument(0).toString().String()
	_key2 := call.Argument(1).toString().String()
	s, err := ro.cli.SDiff(c0.Background(), _key1, _key2).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedisV8_sinter(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.sinter called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key1 := call.Argument(0).toString().String()
	_key2 := call.Argument(1).toString().String()
	s, err := ro.cli.SInter(c0.Background(), _key1, _key2).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedisV8_sisMember(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.sisMember called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_obj := call.Argument(1).Export()
	b, err := ro.cli.SIsMember(c0.Background(), _key, _obj).Result()
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(b)
	}
}

func (r *Runtime) builtinRedisV8_smembers(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.smembers called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.SMembers(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedisV8_smove(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.smove called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}

	_key1 := call.Argument(0).toString().String()
	_key2 := call.Argument(1).toString().String()
	_obj := call.Argument(2).Export()
	b, err := ro.cli.SMove(c0.Background(), _key1, _key2, _obj).Result()
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(b)
	}
}

func (r *Runtime) builtinRedisV8_spop(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.spop called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.SPop(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedisV8_srandMember(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.srandMember called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.SRandMember(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedisV8_srem(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.srem called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_obj := call.Argument(1).Export()
	i, err := ro.cli.SRem(c0.Background(), _key, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_sunion(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.sunion called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key1 := call.Argument(0).toString().String()
	_key2 := call.Argument(1).toString().String()
	s, err := ro.cli.SUnion(c0.Background(), _key1, _key2).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedisV8_hget(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.hget called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_field := call.Argument(1).toString().String()
	s, err := ro.cli.HGet(c0.Background(), _key, _field).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedisV8_hdel(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.hdel called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_field := call.Argument(1).toString().String()
	i, err := ro.cli.HDel(c0.Background(), _key, _field).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_hexists(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.hexists called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_field := call.Argument(1).toString().String()
	b, err := ro.cli.HExists(c0.Background(), _key, _field).Result()
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(b)
	}
}

func (r *Runtime) builtinRedisV8_hgetAll(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.getall called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	m, err := ro.cli.HGetAll(c0.Background(), _key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(m)
	}
}

func (r *Runtime) builtinRedisV8_hkeys(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.hkeys called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	s, err := ro.cli.HKeys(c0.Background(), _key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedisV8_hlen(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.hlen called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	i, err := ro.cli.HLen(c0.Background(), _key).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_hset(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.hset called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_field := call.Argument(1).toString().String()
	_obj := call.Argument(2).Export()
	i, err := ro.cli.HSet(c0.Background(), _key, _field, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {

		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisV8_hvals(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisV8.prototype.hvals called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	s, err := ro.cli.HVals(c0.Background(), _key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtin_newRedisV8(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("RedisV8"))
	}
	if len(args) != 4 {
		panic("number of arguments must be 4")
	}
	_host := args[0].toString().String()
	_port := args[1].ToInteger()
	_password := args[2].toString().String()
	_dbIdx := args[3].ToInteger()
	// 连接 redis
	c := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", _host, _port),
		Password: _password,
		DB:       int(_dbIdx),
	})
	_, err := c.Ping(c0.Background()).Result()
	if err != nil {
		return nil
	} else {
		proto := r.getPrototypeFromCtor(newTarget, r.global.RedisV8, r.global.RedisV8Prototype)
		o := &Object{runtime: r}
		ro := &redisV8Object{
			baseObject: baseObject{
				class:      classRedisV8,
				val:        o,
				prototype:  proto,
				extensible: true,
				values:     nil,
			},
			cli: c,
		}
		o.self = ro
		ro.init()
		return o
	}
}

func (r *Runtime) createRedisV8Proto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("constructor", r.global.RedisV8, true, false, true)
	o._putProp("close", r.newNativeFunc(r.builtinRedisV8_close, nil, "close", nil, 0), true, false, true)

	// string
	o._putProp("get", r.newNativeFunc(r.builtinRedisV8_get, nil, "get", nil, 1), true, false, true)
	o._putProp("set", r.newNativeFunc(r.builtinRedisV8_set, nil, "set", nil, 3), true, false, true)
	o._putProp("del", r.newNativeFunc(r.builtinRedisV8_del, nil, "del", nil, 1), true, false, true)
	o._putProp("ttl", r.newNativeFunc(r.builtinRedisV8_ttl, nil, "ttl", nil, 1), true, false, true)
	// list
	o._putProp("llen", r.newNativeFunc(r.builtinRedisV8_llen, nil, "llen", nil, 1), true, false, true)
	o._putProp("lrange", r.newNativeFunc(r.builtinRedisV8_lrange, nil, "lrange", nil, 3), true, false, true)
	o._putProp("lindex", r.newNativeFunc(r.builtinRedisV8_lindex, nil, "lindex", nil, 2), true, false, true)
	o._putProp("lpush", r.newNativeFunc(r.builtinRedisV8_lpush, nil, "lpush", nil, 2), true, false, true)
	o._putProp("lpop", r.newNativeFunc(r.builtinRedisV8_lpop, nil, "lpop", nil, 1), true, false, true)
	o._putProp("lset", r.newNativeFunc(r.builtinRedisV8_lset, nil, "lset", nil, 3), true, false, true)
	o._putProp("ltrim", r.newNativeFunc(r.builtinRedisV8_ltrim, nil, "ltrim", nil, 3), true, false, true)
	o._putProp("lrem", r.newNativeFunc(r.builtinRedisV8_lrem, nil, "lrem", nil, 3), true, false, true)
	o._putProp("linsertAfter", r.newNativeFunc(r.builtinRedisV8_linsertAfter, nil, "linsertAfter", nil, 3), true, false, true)
	o._putProp("linsertBefore", r.newNativeFunc(r.builtinRedisV8_linsertBefore, nil, "linsertBefore", nil, 3), true, false, true)
	// set
	o._putProp("sadd", r.newNativeFunc(r.builtinRedisV8_sadd, nil, "sadd", nil, 2), true, false, true)
	o._putProp("scard", r.newNativeFunc(r.builtinRedisV8_scard, nil, "scard", nil, 1), true, false, true)
	o._putProp("sdiff", r.newNativeFunc(r.builtinRedisV8_sdiff, nil, "sdiff", nil, 2), true, false, true)
	o._putProp("sinter", r.newNativeFunc(r.builtinRedisV8_sinter, nil, "sinter", nil, 2), true, false, true)
	o._putProp("sisMember", r.newNativeFunc(r.builtinRedisV8_sisMember, nil, "sisMember", nil, 2), true, false, true)
	o._putProp("smembers", r.newNativeFunc(r.builtinRedisV8_smembers, nil, "smembers", nil, 1), true, false, true)
	o._putProp("smove", r.newNativeFunc(r.builtinRedisV8_smove, nil, "smove", nil, 3), true, false, true)
	o._putProp("spop", r.newNativeFunc(r.builtinRedisV8_spop, nil, "spop", nil, 1), true, false, true)
	o._putProp("srandMember", r.newNativeFunc(r.builtinRedisV8_srandMember, nil, "srandMember", nil, 1), true, false, true)
	o._putProp("srem", r.newNativeFunc(r.builtinRedisV8_srem, nil, "srem", nil, 2), true, false, true)
	o._putProp("sunion", r.newNativeFunc(r.builtinRedisV8_sunion, nil, "sunion", nil, 2), true, false, true)
	// hash
	o._putProp("hget", r.newNativeFunc(r.builtinRedisV8_hget, nil, "hget", nil, 2), true, false, true)
	o._putProp("hdel", r.newNativeFunc(r.builtinRedisV8_hdel, nil, "hdel", nil, 2), true, false, true)
	o._putProp("hexists", r.newNativeFunc(r.builtinRedisV8_hexists, nil, "hexists", nil, 2), true, false, true)
	o._putProp("hgetAll", r.newNativeFunc(r.builtinRedisV8_hgetAll, nil, "hgetAll", nil, 1), true, false, true)
	o._putProp("hkeys", r.newNativeFunc(r.builtinRedisV8_hkeys, nil, "hkeys", nil, 1), true, false, true)
	o._putProp("hlen", r.newNativeFunc(r.builtinRedisV8_hlen, nil, "hlen", nil, 1), true, false, true)
	o._putProp("hset", r.newNativeFunc(r.builtinRedisV8_hset, nil, "hset", nil, 3), true, false, true)
	o._putProp("hvals", r.newNativeFunc(r.builtinRedisV8_hvals, nil, "hvals", nil, 1), true, false, true)

	o._putSym(SymToStringTag, valueProp(asciiString(classRedisV8), false, false, true))
	return o
}

func (r *Runtime) createRedisV8(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newRedisV8, r.global.RedisV8Prototype, "RedisV8", 4)
	return o
}

func (r *Runtime) initRedisV8() {
	r.global.RedisV8Prototype = r.newLazyObject(r.createRedisV8Proto)
	r.global.RedisV8 = r.newLazyObject(r.createRedisV8)
	r.addToGlobal("RedisV8", r.global.RedisV8)
}
