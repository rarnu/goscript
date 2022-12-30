package goscript

import (
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type redisObject struct {
	baseObject
	cli *redis.Client
}

func (r *Runtime) builtinRedis_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = ro.cli.Close()
	return _undefined
}

func (r *Runtime) builtinRedis_get(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.get called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.Get(call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedis_set(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.set called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_val := call.Argument(1).Export()
	_expire := call.Argument(2).ToInteger()
	err := ro.cli.Set(_key, _val, time.Millisecond*time.Duration(_expire)).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedis_del(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.del called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	i, err := ro.cli.Del(call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_ttl(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.ttl called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	d, err := ro.cli.TTL(call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(int64(d))
	}
}

func (r *Runtime) builtinRedis_llen(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.llen called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	l, err := ro.cli.LLen(call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	}
	return intToValue(l)
}

func (r *Runtime) builtinRedis_lrange(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.lrange called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_start := call.Argument(1).ToInteger()
	_stop := call.Argument(2).ToInteger()
	s, err := ro.cli.LRange(_key, _start, _stop).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedis_lindex(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.lindex called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_idx := call.Argument(1).ToInteger()
	s, err := ro.cli.LIndex(_key, _idx).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedis_lpush(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.lpush called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_obj := call.Argument(1).Export()
	i, err := ro.cli.LPush(_key, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_lpop(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.lpop called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.LPop(call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedis_lset(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.lset called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_idx := call.Argument(1).ToInteger()
	_obj := call.Argument(2).Export()
	err := ro.cli.LSet(_key, _idx, _obj).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedis_ltrim(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.ltrim called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_start := call.Argument(1).ToInteger()
	_stop := call.Argument(2).ToInteger()
	err := ro.cli.LTrim(_key, _start, _stop).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedis_lrem(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.lrem called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_idx := call.Argument(1).ToInteger()
	_obj := call.Argument(2).Export()
	i, err := ro.cli.LRem(_key, _idx, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_linsertAfter(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.linsertAfter called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_pivot := call.Argument(1).Export()
	_val := call.Argument(2).Export()
	i, err := ro.cli.LInsertAfter(_key, _pivot, _val).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_linsertBefore(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.linsertBefore called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_pivot := call.Argument(1).Export()
	_val := call.Argument(2).Export()
	i, err := ro.cli.LInsertBefore(_key, _pivot, _val).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_sadd(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.sadd called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_obj := call.Argument(1).Export()
	i, err := ro.cli.SAdd(_key, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_scard(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.scard called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	i, err := ro.cli.SCard(call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_sdiff(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.sdiff called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key1 := call.Argument(0).toString().String()
	_key2 := call.Argument(1).toString().String()
	s, err := ro.cli.SDiff(_key1, _key2).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedis_sinter(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.sinter called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key1 := call.Argument(0).toString().String()
	_key2 := call.Argument(1).toString().String()
	s, err := ro.cli.SInter(_key1, _key2).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedis_sisMember(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.sisMember called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_obj := call.Argument(1).Export()
	b, err := ro.cli.SIsMember(_key, _obj).Result()
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(b)
	}
}

func (r *Runtime) builtinRedis_smembers(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.smembers called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.SMembers(call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedis_smove(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.smove called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}

	_key1 := call.Argument(0).toString().String()
	_key2 := call.Argument(1).toString().String()
	_obj := call.Argument(2).Export()
	b, err := ro.cli.SMove(_key1, _key2, _obj).Result()
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(b)
	}
}

func (r *Runtime) builtinRedis_spop(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.spop called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.SPop(call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedis_srandMember(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.srandMember called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.SRandMember(call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedis_srem(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.srem called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_obj := call.Argument(1).Export()
	i, err := ro.cli.SRem(_key, _obj).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_sunion(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.sunion called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key1 := call.Argument(0).toString().String()
	_key2 := call.Argument(1).toString().String()
	s, err := ro.cli.SUnion(_key1, _key2).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedis_hget(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.hget called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_field := call.Argument(1).toString().String()
	s, err := ro.cli.HGet(_key, _field).Result()
	if err != nil {
		return _null
	} else {
		return &importedString{
			s: s,
		}
	}
}

func (r *Runtime) builtinRedis_hdel(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.hdel called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_field := call.Argument(1).toString().String()
	i, err := ro.cli.HDel(_key, _field).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_hexists(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.hexists called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_field := call.Argument(1).toString().String()
	b, err := ro.cli.HExists(_key, _field).Result()
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(b)
	}
}

func (r *Runtime) builtinRedis_hgetAll(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.getall called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	m, err := ro.cli.HGetAll(_key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(m)
	}
}

func (r *Runtime) builtinRedis_hkeys(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.hkeys called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	s, err := ro.cli.HKeys(_key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedis_hlen(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.hlen called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	i, err := ro.cli.HLen(_key).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedis_hset(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.hset called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_field := call.Argument(1).toString().String()
	_obj := call.Argument(2).Export()
	b, err := ro.cli.HSet(_key, _field, _obj).Result()
	if err != nil {
		return valueFalse
	} else {
		return r.toBoolean(b)
	}
}

func (r *Runtime) builtinRedis_hvals(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisObject)
	if !ok {
		panic(r.NewTypeError("Method Redis.prototype.hvals called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	s, err := ro.cli.HVals(_key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtin_newRedis(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("Redis"))
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
	_, err := c.Ping().Result()
	if err != nil {
		return nil
	} else {
		proto := r.getPrototypeFromCtor(newTarget, r.global.Redis, r.global.RedisPrototype)
		o := &Object{runtime: r}
		ro := &redisObject{
			baseObject: baseObject{
				class:      classRedis,
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

func (r *Runtime) createRedisProto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("constructor", r.global.Redis, true, false, true)
	o._putProp("close", r.newNativeFunc(r.builtinRedis_close, nil, "close", nil, 0), true, false, true)

	// string
	o._putProp("get", r.newNativeFunc(r.builtinRedis_get, nil, "get", nil, 1), true, false, true)
	o._putProp("set", r.newNativeFunc(r.builtinRedis_set, nil, "set", nil, 3), true, false, true)
	o._putProp("del", r.newNativeFunc(r.builtinRedis_del, nil, "del", nil, 1), true, false, true)
	o._putProp("ttl", r.newNativeFunc(r.builtinRedis_ttl, nil, "ttl", nil, 1), true, false, true)
	// list
	o._putProp("llen", r.newNativeFunc(r.builtinRedis_llen, nil, "llen", nil, 1), true, false, true)
	o._putProp("lrange", r.newNativeFunc(r.builtinRedis_lrange, nil, "lrange", nil, 3), true, false, true)
	o._putProp("lindex", r.newNativeFunc(r.builtinRedis_lindex, nil, "lindex", nil, 2), true, false, true)
	o._putProp("lpush", r.newNativeFunc(r.builtinRedis_lpush, nil, "lpush", nil, 2), true, false, true)
	o._putProp("lpop", r.newNativeFunc(r.builtinRedis_lpop, nil, "lpop", nil, 1), true, false, true)
	o._putProp("lset", r.newNativeFunc(r.builtinRedis_lset, nil, "lset", nil, 3), true, false, true)
	o._putProp("ltrim", r.newNativeFunc(r.builtinRedis_ltrim, nil, "ltrim", nil, 3), true, false, true)
	o._putProp("lrem", r.newNativeFunc(r.builtinRedis_lrem, nil, "lrem", nil, 3), true, false, true)
	o._putProp("linsertAfter", r.newNativeFunc(r.builtinRedis_linsertAfter, nil, "linsertAfter", nil, 3), true, false, true)
	o._putProp("linsertBefore", r.newNativeFunc(r.builtinRedis_linsertBefore, nil, "linsertBefore", nil, 3), true, false, true)
	// set
	o._putProp("sadd", r.newNativeFunc(r.builtinRedis_sadd, nil, "sadd", nil, 2), true, false, true)
	o._putProp("scard", r.newNativeFunc(r.builtinRedis_scard, nil, "scard", nil, 1), true, false, true)
	o._putProp("sdiff", r.newNativeFunc(r.builtinRedis_sdiff, nil, "sdiff", nil, 2), true, false, true)
	o._putProp("sinter", r.newNativeFunc(r.builtinRedis_sinter, nil, "sinter", nil, 2), true, false, true)
	o._putProp("sisMember", r.newNativeFunc(r.builtinRedis_sisMember, nil, "sisMember", nil, 2), true, false, true)
	o._putProp("smembers", r.newNativeFunc(r.builtinRedis_smembers, nil, "smembers", nil, 1), true, false, true)
	o._putProp("smove", r.newNativeFunc(r.builtinRedis_smove, nil, "smove", nil, 3), true, false, true)
	o._putProp("spop", r.newNativeFunc(r.builtinRedis_spop, nil, "spop", nil, 1), true, false, true)
	o._putProp("srandMember", r.newNativeFunc(r.builtinRedis_srandMember, nil, "srandMember", nil, 1), true, false, true)
	o._putProp("srem", r.newNativeFunc(r.builtinRedis_srem, nil, "srem", nil, 2), true, false, true)
	o._putProp("sunion", r.newNativeFunc(r.builtinRedis_sunion, nil, "sunion", nil, 2), true, false, true)
	// hash
	o._putProp("hget", r.newNativeFunc(r.builtinRedis_hget, nil, "hget", nil, 2), true, false, true)
	o._putProp("hdel", r.newNativeFunc(r.builtinRedis_hdel, nil, "hdel", nil, 2), true, false, true)
	o._putProp("hexists", r.newNativeFunc(r.builtinRedis_hexists, nil, "hexists", nil, 2), true, false, true)
	o._putProp("hgetAll", r.newNativeFunc(r.builtinRedis_hgetAll, nil, "hgetAll", nil, 1), true, false, true)
	o._putProp("hkeys", r.newNativeFunc(r.builtinRedis_hkeys, nil, "hkeys", nil, 1), true, false, true)
	o._putProp("hlen", r.newNativeFunc(r.builtinRedis_hlen, nil, "hlen", nil, 1), true, false, true)
	o._putProp("hset", r.newNativeFunc(r.builtinRedis_hset, nil, "hset", nil, 3), true, false, true)
	o._putProp("hvals", r.newNativeFunc(r.builtinRedis_hvals, nil, "hvals", nil, 1), true, false, true)

	o._putSym(SymToStringTag, valueProp(asciiString(classRedis), false, false, true))
	return o
}

func (r *Runtime) createRedis(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newRedis, r.global.RedisPrototype, "Redis", 4)
	return o
}

func (r *Runtime) initRedis() {
	r.global.RedisPrototype = r.newLazyObject(r.createRedisProto)
	r.global.Redis = r.newLazyObject(r.createRedis)
	r.addToGlobal("Redis", r.global.Redis)
}
