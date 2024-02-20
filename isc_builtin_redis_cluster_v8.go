package goscript

import (
	c0 "context"
	"github.com/go-redis/redis/v8"
	"runtime"
	"strings"
	"time"
)

type redisClusterV8Object struct {
	baseObject
	cli *redis.ClusterClient
}

func (r *Runtime) builtinRedisClusterV8_close(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.close called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_ = ro.cli.Close()
	return _undefined
}

func (r *Runtime) builtinRedisClusterV8_get(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.get called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_set(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.set called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_val := call.Argument(1).Export()
	_expire := call.Argument(2).ToInteger()
	err := ro.cli.Set(c0.Background(), _key, _val, time.Millisecond*time.Duration(_expire)).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedisClusterV8_del(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.del called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	i, err := ro.cli.Del(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisClusterV8_ttl(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.ttl called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	d, err := ro.cli.TTL(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(int64(d))
	}
}

func (r *Runtime) builtinRedisClusterV8_llen(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.llen called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	l, err := ro.cli.LLen(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	}
	return intToValue(l)
}

func (r *Runtime) builtinRedisClusterV8_lrange(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.lrange called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_lindex(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.lindex called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_lpush(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.lpush called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_lpop(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.lpop called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_lset(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.lset called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_idx := call.Argument(1).ToInteger()
	_obj := call.Argument(2).Export()
	err := ro.cli.LSet(c0.Background(), _key, _idx, _obj).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedisClusterV8_ltrim(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.ltrim called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	_start := call.Argument(1).ToInteger()
	_stop := call.Argument(2).ToInteger()
	err := ro.cli.LTrim(c0.Background(), _key, _start, _stop).Err()
	return r.toBoolean(err == nil)
}

func (r *Runtime) builtinRedisClusterV8_lrem(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.lrem called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_linsertAfter(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.linsertAfter called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_linsertBefore(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.linsertBefore called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_sadd(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.sadd called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_scard(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.scard called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	i, err := ro.cli.SCard(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisClusterV8_sdiff(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.sdiff called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_sinter(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.sinter called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_sisMember(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.sisMember called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_smembers(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.smembers called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	s, err := ro.cli.SMembers(c0.Background(), call.Argument(0).toString().String()).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedisClusterV8_smove(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.smove called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_spop(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.spop called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_srandMember(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.srandMember called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_srem(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.srem called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_sunion(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.sunion called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_hget(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.hget called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_hdel(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.hdel called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_hexists(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.hexists called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_hgetAll(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.getall called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	m, err := ro.cli.HGetAll(c0.Background(), _key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(m)
	}
}

func (r *Runtime) builtinRedisClusterV8_hkeys(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.hkeys called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	s, err := ro.cli.HKeys(c0.Background(), _key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtinRedisClusterV8_hlen(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.hlen called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	i, err := ro.cli.HLen(c0.Background(), _key).Result()
	if err != nil {
		return intToValue(-1)
	} else {
		return intToValue(i)
	}
}

func (r *Runtime) builtinRedisClusterV8_hset(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.hset called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
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

func (r *Runtime) builtinRedisClusterV8_hvals(call FunctionCall) Value {
	thisObj := r.toObject(call.This)
	ro, ok := thisObj.self.(*redisClusterV8Object)
	if !ok {
		panic(r.NewTypeError("Method RedisClusterV8.prototype.hvals called on incompatible receiver %s", r.objectproto_toString(FunctionCall{This: thisObj})))
	}
	_key := call.Argument(0).toString().String()
	s, err := ro.cli.HVals(c0.Background(), _key).Result()
	if err != nil {
		return _null
	} else {
		return r.ToValue(s)
	}
}

func (r *Runtime) builtin_newRedisClusterV8(args []Value, newTarget *Object) *Object {
	if newTarget == nil {
		panic(r.needNew("RedisClusterV8"))
	}
	if len(args) != 2 {
		panic("number of arguments must be 2")
	}

	_address := args[0].toString().String()
	addrs := strings.Split(_address, ";")
	_password := args[1].toString().String()

	// 连接 redis 集群
	c := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:        addrs,
		Password:     _password,
		ReadOnly:     false,
		PoolSize:     20 * runtime.NumCPU(),
		MinIdleConns: 10,
	})
	_, err := c.Ping(c0.Background()).Result()
	if err != nil {
		return nil
	} else {
		proto := r.getPrototypeFromCtor(newTarget, r.global.RedisClusterV8, r.global.RedisClusterV8Prototype)
		o := &Object{runtime: r}
		ro := &redisClusterV8Object{
			baseObject: baseObject{
				class:      classRedisClusterV8,
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

func (r *Runtime) createRedisClusterV8Proto(val *Object) objectImpl {
	o := newBaseObjectObj(val, r.global.ObjectPrototype, classObject)
	o._putProp("constructor", r.global.RedisClusterV8, true, false, true)
	o._putProp("close", r.newNativeFunc(r.builtinRedisClusterV8_close, nil, "close", nil, 0), true, false, true)

	// string
	o._putProp("get", r.newNativeFunc(r.builtinRedisClusterV8_get, nil, "get", nil, 1), true, false, true)
	o._putProp("set", r.newNativeFunc(r.builtinRedisClusterV8_set, nil, "set", nil, 3), true, false, true)
	o._putProp("del", r.newNativeFunc(r.builtinRedisClusterV8_del, nil, "del", nil, 1), true, false, true)
	o._putProp("ttl", r.newNativeFunc(r.builtinRedisClusterV8_ttl, nil, "ttl", nil, 1), true, false, true)
	// list
	o._putProp("llen", r.newNativeFunc(r.builtinRedisClusterV8_llen, nil, "llen", nil, 1), true, false, true)
	o._putProp("lrange", r.newNativeFunc(r.builtinRedisClusterV8_lrange, nil, "lrange", nil, 3), true, false, true)
	o._putProp("lindex", r.newNativeFunc(r.builtinRedisClusterV8_lindex, nil, "lindex", nil, 2), true, false, true)
	o._putProp("lpush", r.newNativeFunc(r.builtinRedisClusterV8_lpush, nil, "lpush", nil, 2), true, false, true)
	o._putProp("lpop", r.newNativeFunc(r.builtinRedisClusterV8_lpop, nil, "lpop", nil, 1), true, false, true)
	o._putProp("lset", r.newNativeFunc(r.builtinRedisClusterV8_lset, nil, "lset", nil, 3), true, false, true)
	o._putProp("ltrim", r.newNativeFunc(r.builtinRedisClusterV8_ltrim, nil, "ltrim", nil, 3), true, false, true)
	o._putProp("lrem", r.newNativeFunc(r.builtinRedisClusterV8_lrem, nil, "lrem", nil, 3), true, false, true)
	o._putProp("linsertAfter", r.newNativeFunc(r.builtinRedisClusterV8_linsertAfter, nil, "linsertAfter", nil, 3), true, false, true)
	o._putProp("linsertBefore", r.newNativeFunc(r.builtinRedisClusterV8_linsertBefore, nil, "linsertBefore", nil, 3), true, false, true)
	// set
	o._putProp("sadd", r.newNativeFunc(r.builtinRedisClusterV8_sadd, nil, "sadd", nil, 2), true, false, true)
	o._putProp("scard", r.newNativeFunc(r.builtinRedisClusterV8_scard, nil, "scard", nil, 1), true, false, true)
	o._putProp("sdiff", r.newNativeFunc(r.builtinRedisClusterV8_sdiff, nil, "sdiff", nil, 2), true, false, true)
	o._putProp("sinter", r.newNativeFunc(r.builtinRedisClusterV8_sinter, nil, "sinter", nil, 2), true, false, true)
	o._putProp("sisMember", r.newNativeFunc(r.builtinRedisClusterV8_sisMember, nil, "sisMember", nil, 2), true, false, true)
	o._putProp("smembers", r.newNativeFunc(r.builtinRedisClusterV8_smembers, nil, "smembers", nil, 1), true, false, true)
	o._putProp("smove", r.newNativeFunc(r.builtinRedisClusterV8_smove, nil, "smove", nil, 3), true, false, true)
	o._putProp("spop", r.newNativeFunc(r.builtinRedisClusterV8_spop, nil, "spop", nil, 1), true, false, true)
	o._putProp("srandMember", r.newNativeFunc(r.builtinRedisClusterV8_srandMember, nil, "srandMember", nil, 1), true, false, true)
	o._putProp("srem", r.newNativeFunc(r.builtinRedisClusterV8_srem, nil, "srem", nil, 2), true, false, true)
	o._putProp("sunion", r.newNativeFunc(r.builtinRedisClusterV8_sunion, nil, "sunion", nil, 2), true, false, true)
	// hash
	o._putProp("hget", r.newNativeFunc(r.builtinRedisClusterV8_hget, nil, "hget", nil, 2), true, false, true)
	o._putProp("hdel", r.newNativeFunc(r.builtinRedisClusterV8_hdel, nil, "hdel", nil, 2), true, false, true)
	o._putProp("hexists", r.newNativeFunc(r.builtinRedisClusterV8_hexists, nil, "hexists", nil, 2), true, false, true)
	o._putProp("hgetAll", r.newNativeFunc(r.builtinRedisClusterV8_hgetAll, nil, "hgetAll", nil, 1), true, false, true)
	o._putProp("hkeys", r.newNativeFunc(r.builtinRedisClusterV8_hkeys, nil, "hkeys", nil, 1), true, false, true)
	o._putProp("hlen", r.newNativeFunc(r.builtinRedisClusterV8_hlen, nil, "hlen", nil, 1), true, false, true)
	o._putProp("hset", r.newNativeFunc(r.builtinRedisClusterV8_hset, nil, "hset", nil, 3), true, false, true)
	o._putProp("hvals", r.newNativeFunc(r.builtinRedisClusterV8_hvals, nil, "hvals", nil, 1), true, false, true)

	o._putSym(SymToStringTag, valueProp(asciiString(classRedisClusterV8), false, false, true))
	return o
}

func (r *Runtime) createRedisClusterV8(val *Object) objectImpl {
	o := r.newNativeConstructOnly(val, r.builtin_newRedisClusterV8, r.global.RedisClusterV8Prototype, "RedisClusterV8", 2)
	return o
}

func (r *Runtime) initRedisClusterV8() {
	r.global.RedisClusterV8Prototype = r.newLazyObject(r.createRedisClusterV8Proto)
	r.global.RedisClusterV8 = r.newLazyObject(r.createRedisClusterV8)
	r.addToGlobal("RedisClusterV8", r.global.RedisClusterV8)
}
