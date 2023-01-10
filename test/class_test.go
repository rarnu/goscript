package test

import (
	"fmt"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/console"
	"github.com/rarnu/goscript/module/require"
	"testing"
)

func TestClassDef(t *testing.T) {
	SCRIPT := `
class Person {
	constructor(name, age) {
		this.name = name
		this.age = age
	}
	foo() {
		console.log(this.name)
	}
}
let p1 = new Person('rarnu', 35)
console.log(p1)
p1.foo()
`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	v, e := vm.RunString(SCRIPT)
	fmt.Printf("v = %+v\n, e = %v\n", v.Export(), e)

}
