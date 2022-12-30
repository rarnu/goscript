package process

import (
	"os"
	"strings"

	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/require"
)

type Process struct {
	env map[string]string
}

func Require(runtime *goscript.Runtime, module *goscript.Object) {
	p := &Process{
		env: make(map[string]string),
	}

	for _, e := range os.Environ() {
		envKeyValue := strings.SplitN(e, "=", 2)
		p.env[envKeyValue[0]] = envKeyValue[1]
	}

	o := module.Get("exports").(*goscript.Object)
	o.Set("env", p.env)
}

func Enable(runtime *goscript.Runtime) {
	runtime.Set("process", require.Require(runtime, "process"))
}

func init() {
	require.RegisterNativeModule("process", Require)
}
