package util

import (
	"bytes"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/require"
)

type Util struct {
	runtime *goscript.Runtime
}

func (u *Util) format(f rune, val goscript.Value, w *bytes.Buffer) bool {
	switch f {
	case 's':
		w.WriteString(val.String())
	case 'd':
		w.WriteString(val.ToNumber().String())
	case 'j':
		if json, ok := u.runtime.Get("JSON").(*goscript.Object); ok {
			if stringify, ok := goscript.AssertFunction(json.Get("stringify")); ok {
				res, err := stringify(json, val)
				if err != nil {
					panic(err)
				}
				w.WriteString(res.String())
			}
		}
	case '%':
		w.WriteByte('%')
		return false
	default:
		w.WriteByte('%')
		w.WriteRune(f)
		return false
	}
	return true
}

func (u *Util) Format(b *bytes.Buffer, f string, args ...goscript.Value) {
	pct := false
	argNum := 0
	for _, chr := range f {
		if pct {
			if argNum < len(args) {
				if u.format(chr, args[argNum], b) {
					argNum++
				}
			} else {
				b.WriteByte('%')
				b.WriteRune(chr)
			}
			pct = false
		} else {
			if chr == '%' {
				pct = true
			} else {
				b.WriteRune(chr)
			}
		}
	}

	for _, arg := range args[argNum:] {
		b.WriteByte(' ')
		b.WriteString(arg.String())
	}
}

func (u *Util) js_format(call goscript.FunctionCall) goscript.Value {
	var b bytes.Buffer
	var fmt string

	if arg := call.Argument(0); !goscript.IsUndefined(arg) {
		fmt = arg.String()
	}

	var args []goscript.Value
	if len(call.Arguments) > 0 {
		args = call.Arguments[1:]
	}
	u.Format(&b, fmt, args...)

	return u.runtime.ToValue(b.String())
}

func Require(runtime *goscript.Runtime, module *goscript.Object) {
	u := &Util{
		runtime: runtime,
	}
	obj := module.Get("exports").(*goscript.Object)
	_ = obj.Set("format", u.js_format)
}

func New(runtime *goscript.Runtime) *Util {
	return &Util{
		runtime: runtime,
	}
}

func init() {
	require.RegisterNativeModule("util", Require)
}
