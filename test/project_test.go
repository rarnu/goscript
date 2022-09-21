package test

import (
	"fmt"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/console"
	"github.com/rarnu/goscript/require"
	"reflect"
	"testing"
)

// TestMapInject map 注入脚本中
func TestMapInject(t *testing.T) {

	SCRIPT := `
console.log($m["a"])
console.log($m["b"])
$m["a"] = 2
$m["b"] = "zzz"
$m.c.d = 666
$m.c.e = 777
`

	m := map[string]any{
		"a": 1,
		"b": "x",
		"c": map[string]any{
			"d": 6,
			"e": 7,
		},
	}

	t.Logf("m-before: %+v\n", m)

	vm := goscript.New()
	// 必须开启允许 require，然后加载 console 模块
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	_ = vm.Set("$m", &m)
	p, _ := goscript.Compile("test.js", SCRIPT, false)
	_, _ = vm.RunProgram(p)
	t.Logf("m-after: %+v\n", m)
}

func TestResultMap(t *testing.T) {
	// 要求，在以上结果集中，去除 name 为 c 的项，并且把剩余的数据项的 val 翻倍
	m := map[string]any{
		"$x.$resp": map[string]any{
			"info": []map[string]any{
				{"name": "a", "val": 1},
				{"name": "b", "val": 2},
				{"name": "c", "val": 3},
			},
		},
	}
	t.Logf("before-changed: %+v\n", m)
	SCRIPT := `
let info = $req['$x.$resp'].info
let filtered = info.filter(it => {
	return it.name !== 'c'
}).map(it => {
	it.val *= 2
	return it
})
$req['$x.$resp'].info = filtered
`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	_ = vm.Set("$req", &m)
	p, _ := goscript.Compile("test.js", SCRIPT, false)
	_, _ = vm.RunProgram(p)
	t.Logf("after-changed: %+v\n", m)
}

func TestFuncInject(t *testing.T) {
	SCRIPT := `
let n = "rarnu"
n = hello(n)
console.log(n)
`
	f := func(str string) string {
		return "hello " + str
	}
	vm := goscript.New()
	// 必须开启允许 require，然后加载 console 模块
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	_ = vm.Set("hello", f)
	p, _ := goscript.Compile("test.js", SCRIPT, false)
	_, _ = vm.RunProgram(p)
}

func TestReturnInject(t *testing.T) {
	SCRIPT := `
let a = "abcdefg"
console.log(a)
a // 最后给一个表达式即可，不需要 return 关键字 
`
	vm := goscript.New()
	// 必须开启允许 require，然后加载 console 模块
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	p, _ := goscript.Compile("test.js", SCRIPT, false)
	v, _ := vm.RunProgram(p)
	t.Logf("return: %s\n", v.ToString())
}

func TestObjectInject(t *testing.T) {
	SCRIPT := `
let a = {"a":1, "b": "xx"}
a
`
	vm := goscript.New()
	p, _ := goscript.Compile("test.js", SCRIPT, false)
	v, _ := vm.RunProgram(p)
	e := v.Export().(map[string]any)
	t.Logf("return: %+v\n", e)

	obj := map[string]any{
		"f1": 1,
		"f2": "zzz",
	}

	SCRIPT1 := `
console.log(obj)
console.log(obj.f1)
console.log(obj.f2)
obj.f1 = 666
obj.f2 = "yyy"
`

	vm1 := goscript.New()
	new(require.Registry).Enable(vm1)
	console.Enable(vm1)
	_ = vm1.Set("obj", &obj)
	p1, _ := goscript.Compile("test.js", SCRIPT1, false)
	_, _ = vm1.RunProgram(p1)
	t.Logf("after-object-changed: %+v\n", obj)

}

func TestArrayInject(t *testing.T) {
	SCRIPT := `
console.log($arr.length)
$arr.push(666)
for (i in $arr) {
	console.log($arr[i])
}
`
	arr := []any{1, 2, 3, 4, 5}

	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	_ = vm.Set("$arr", &arr)
	p, _ := goscript.Compile("test.js", SCRIPT, false)
	_, _ = vm.RunProgram(p)
	t.Logf("after-array-changed: %+v\n", arr)
}

func TestMathAlgorithms(t *testing.T) {
	SCRIPT := `
function s(k) {
	return 2 * Math.sin(k)
}
s(3)
`
	vm := goscript.New()
	p, _ := goscript.Compile("test.js", SCRIPT, false)
	v, err := vm.RunProgram(p)
	t.Logf("v = %+v\n, err = %v\n", v, err)
}

func TestRequire(t *testing.T) {
	SCRIPT := `
const m = require("../test/md5.js")
m.md5('abcdefg')
`
	vm := goscript.New()

	registry := new(require.Registry)
	registry.Enable(vm)
	console.Enable(vm)

	p, _ := goscript.Compile("test.js", SCRIPT, false)
	v, _ := vm.RunProgram(p)
	if v.String() != "7ac66c0f148de9519b8bd264312c4d64" {
		t.Fatalf("md5 failed")
	}
}

func TestCrypto(t *testing.T) {
	SCRIPT := `
let a = Crypto.md5('abcdefg')
console.log(a)
a
`
	vm := goscript.New()

	registry := new(require.Registry)
	registry.Enable(vm)
	console.Enable(vm)

	p, _ := goscript.Compile("test.js", SCRIPT, false)
	v, _ := vm.RunProgram(p)
	if v.String() != "7ac66c0f148de9519b8bd264312c4d64" {
		t.Fatalf("md5 failed")
	}
}

func TestHttp(t *testing.T) {
	SCRIPT := `
let ret = HTTP.get('http://10.30.30.78:29013/api/core/license/info', {"token": "abcdefg"})
console.log(ret.statusCode)
console.log(ret.data.data.licenseCode)
console.log(ret.data.data.customer.enterpriseName)
`
	vm := goscript.New()

	registry := new(require.Registry)
	registry.Enable(vm)
	console.Enable(vm)
	p, _ := goscript.Compile("test.js", SCRIPT, false)
	_, _ = vm.RunProgram(p)
}

func TestMysql(t *testing.T) {
	SCRIPT := `
let db = new Mysql('10.30.30.81', 23306, 'isyscore', 'Isysc0re', 'isc_service')
let rows = db.query('select * from isc_env limit 0, 10')
for (i in rows) {
    console.log(rows[i].name)
	console.log(rows[i].domain)
}
db.close()
`
	vm := goscript.New()
	registry := new(require.Registry)
	registry.Enable(vm)
	console.Enable(vm)
	p, err := goscript.Compile("test.js", SCRIPT, false)
	t.Logf("compile error: %v", err)
	_, err = vm.RunProgram(p)
	t.Logf("run error: %v", err)

}

func TestRedis(t *testing.T) {
	SCRIPT := `
let cli = new Redis('10.30.30.81', 26379, 'ZljIsysc0re123', 7)
cli.set('go-script:sample-id', 'abcdefg')
let i = cli.get('go-script:sample-id')
console.log(i)
cli.close()
`
	vm := goscript.New()
	registry := new(require.Registry)
	registry.Enable(vm)
	console.Enable(vm)
	p, err := goscript.Compile("test.js", SCRIPT, false)
	t.Logf("compile error: %v", err)
	_, err = vm.RunProgram(p)
	t.Logf("run error: %v", err)

}

func TestRunCode(t *testing.T) {
	SCRIPT := `
	console.log('sample')
	`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	v0, err := vm.RunString(SCRIPT)
	vx := v0.Export()
	t.Logf("v0 = %+v\n, err = %+v\n", vx, err)
}

func TestExportMap(t *testing.T) {
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	m, err := vm.RunString(`let a = new Map()
a.set(1, true)
a.set(2, false)
a
// new Map([[1, true], [2, false]]);
	`)
	if err != nil {
		panic(err)
	}
	exp := m.Export()
	_typ := reflect.TypeOf(exp).String()
	fmt.Printf("typ = %s\n", _typ)

	fmt.Printf("%T, %v\n", exp, exp)
	// 输出: [][2]interface {}, [[1 true] [2 false]]
}

func TestExportObject(t *testing.T) {
	SCRIPT := `let obj = {id:1, name: 'rarnu', age: 35}
console.log(obj)
obj
`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	v, e := vm.RunString(SCRIPT)

	fmt.Printf("v = %+v\n, e = %v\n", v.Export(), e)

}

func TestExportSimpleObject(t *testing.T) {
	SCRIPT := `let obj = {id:1, name: 'rarnu', age: 35}
obj
`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	v, e := vm.RunString(SCRIPT)

	fmt.Printf("v = %+v\n, e = %v\n", v.Export(), e)

}
