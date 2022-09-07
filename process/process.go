package process

import (
	p0 "github.com/isyscore/isc-gobase/system/process"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/require"
	"os"
	"strings"
)

type Process struct {
	runtime *goscript.Runtime
	env     map[string]string
}

func (p *Process) list(call goscript.FunctionCall) goscript.Value {
	ps, err := p0.Processes()
	if err != nil {
		return goscript.Null()
	} else {
		var ret0 []map[string]any
		for _, item := range ps {
			_p0, _ := item.Name()
			_p1, _ := item.Status()
			_p2, _ := item.Exe()
			_p3, _ := item.Cmdline()
			_p4, _ := item.Uids()
			_p5, _ := item.Gids()
			_p6, _ := item.CPUPercent()
			_p7, _ := item.MemoryPercent()
			_p8, _ := item.Username()
			_m := map[string]any{
				"pid":        item.Pid,
				"name":       _p0,
				"status":     _p1,
				"exe":        _p2,
				"cmdline":    _p3,
				"uids":       _p4,
				"gids":       _p5,
				"cpuPercent": _p6,
				"memPercent": _p7,
				"user":       _p8,
			}
			ret0 = append(ret0, _m)
		}
		return p.runtime.ToValue(ret0)
	}
}

func (p *Process) kill(call goscript.FunctionCall) goscript.Value {
	proc, err := p0.NewProcess(int32(call.Argument(0).ToInteger()))
	if err != nil {
		return goscript.False()
	}
	err = proc.Kill()
	return p.runtime.ToValue(err == nil)
}

func Require(runtime *goscript.Runtime, module *goscript.Object) {
	p := &Process{
		runtime: runtime,
		env:     make(map[string]string),
	}

	for _, e := range os.Environ() {
		envKeyValue := strings.SplitN(e, "=", 2)
		p.env[envKeyValue[0]] = envKeyValue[1]
	}

	o := module.Get("exports").(*goscript.Object)
	_ = o.Set("env", p.env)
	_ = o.Set("list", p.list)
	_ = o.Set("kill", p.kill)
}

func Enable(runtime *goscript.Runtime) {
	_ = runtime.Set("process", require.Require(runtime, "process"))
}

func init() {
	require.RegisterNativeModule("process", Require)
}
