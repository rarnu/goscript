package test

import (
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/console"
	"github.com/rarnu/goscript/module/require"
	"testing"
)

func TestRedisCluster(t *testing.T) {
	SCRIPT := `
let cli = new RedisCluster('10.211.55.5:6379', 'rootroot')
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

func TestRedis(t *testing.T) {
	SCRIPT := `
let cli = new Redis('10.211.55.5', 6379, 'rootroot', 0)
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

func TestRedisV8(t *testing.T) {
	SCRIPT := `
let cli = new RedisV8('10.211.55.5', 6379, 'rootroot', 0)
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

func TestRedisClusterV8(t *testing.T) {
	SCRIPT := `
let cli = new RedisClusterV8('10.211.55.5:6379', 'rootroot')
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
