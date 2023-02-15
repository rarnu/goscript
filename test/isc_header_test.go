package test

import (
	"fmt"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/console"
	"github.com/rarnu/goscript/module/require"
	"testing"
)

func TestReadHeader(t *testing.T) {
	SCRIPT := `
	console.log('token = ' + $req.header.token)
	$req.header.token1 = '66666'
	`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	m := map[string]any{
		"header": map[string]any{
			"Token":  "abcde",
			"Token1": "88888",
		},
	}
	_ = vm.Set("$req", &m)
	_, _ = vm.RunString(SCRIPT)
	fmt.Printf("m: %+v\n", m)
}
