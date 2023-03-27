package test

import (
	"fmt"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/console"
	"github.com/rarnu/goscript/module/require"
	"testing"
)

func TestExport(t *testing.T) {

	SCRIPT := `
// 1. 操作数据
let a = $loopItem;
let b = a + 1;

// 2. 设置返回值

const result = {
    a: a,
    b: b
};

// 3. 确认返回值
result;
`

	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)

	var item any = 1
	_ = vm.Set("$loopItem", item)
	v, e := vm.RunString(SCRIPT)
	vx0 := v.Export().(map[string]any)
	a := vx0["a"]
	b := vx0["b"]
	fmt.Printf("a = %d, b = %d\n", a, b)
	fmt.Printf("v = %+v\n, e = %v\n, vx0 = %+v\n", v.Export(), e, vx0)
}
