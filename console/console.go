package console

import (
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/require"
	_ "github.com/rarnu/goscript/util"
	"log"
)

type Console struct {
	runtime *goscript.Runtime
	util    *goscript.Object
	printer Printer
}

type Printer interface {
	Log(string)
	Warn(string)
	Error(string)
}

type PrinterFunc func(s string)

func (p PrinterFunc) Log(s string) { p(s) }

func (p PrinterFunc) Warn(s string) { p(s) }

func (p PrinterFunc) Error(s string) { p(s) }

var defaultPrinter Printer = PrinterFunc(func(s string) { log.Print(s) })

func (c *Console) log(p func(string)) func(goscript.FunctionCall) goscript.Value {
	return func(call goscript.FunctionCall) goscript.Value {
		if format, ok := goscript.AssertFunction(c.util.Get("format")); ok {
			ret, err := format(c.util, call.Arguments...)
			if err != nil {
				panic(err)
			}

			p(ret.String())
		} else {
			panic(c.runtime.NewTypeError("util.format is not a function"))
		}
		return nil
	}
}

func Require(runtime *goscript.Runtime, module *goscript.Object) {
	requireWithPrinter(defaultPrinter)(runtime, module)
}

func RequireWithPrinter(printer Printer) require.ModuleLoader {
	return requireWithPrinter(printer)
}

func requireWithPrinter(printer Printer) require.ModuleLoader {
	return func(runtime *goscript.Runtime, module *goscript.Object) {
		c := &Console{
			runtime: runtime,
			printer: printer,
		}

		c.util = require.Require(runtime, "util").(*goscript.Object)

		o := module.Get("exports").(*goscript.Object)
		_ = o.Set("log", c.log(c.printer.Log))
		_ = o.Set("error", c.log(c.printer.Error))
		_ = o.Set("warn", c.log(c.printer.Warn))
	}
}

func Enable(runtime *goscript.Runtime) {
	_ = runtime.Set("console", require.Require(runtime, "console"))
}

func init() {
	require.RegisterNativeModule("console", Require)
}
