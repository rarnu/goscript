package console

import (
	"github.com/rarnu/goscript"
	"log"

	"github.com/rarnu/goscript/module/require"
	"github.com/rarnu/goscript/module/util"
)

const ModuleName = "node:console"

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

		c.util = require.Require(runtime, util.ModuleName).(*goscript.Object)

		o := module.Get("exports").(*goscript.Object)
		o.Set("log", c.log(c.printer.Log))
		o.Set("error", c.log(c.printer.Error))
		o.Set("warn", c.log(c.printer.Warn))
	}
}

func Enable(runtime *goscript.Runtime) {
	runtime.Set("console", require.Require(runtime, ModuleName))
}

func init() {
	require.RegisterNativeModule(ModuleName, Require)
}
