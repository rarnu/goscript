package main

import (
	crand "crypto/rand"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/console"
	"github.com/rarnu/goscript/require"
	"io"
	"log"
	"math/rand"
	"os"
	"runtime/debug"
	"runtime/pprof"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
var timelimit = flag.Int("timelimit", 0, "max time to run (in seconds)")

func readSource(filename string) ([]byte, error) {
	if filename == "" || filename == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(filename)
}

func load(vm *goscript.Runtime, call goscript.FunctionCall) goscript.Value {
	p := call.Argument(0).String()
	b, err := readSource(p)
	if err != nil {
		panic(vm.ToValue(fmt.Sprintf("Could not read %s: %v", p, err)))
	}
	v, err := vm.RunScript(p, string(b))
	if err != nil {
		panic(err)
	}
	return v
}

func newRandSource() goscript.RandSource {
	var seed int64
	if err := binary.Read(crand.Reader, binary.LittleEndian, &seed); err != nil {
		panic(fmt.Errorf("could not read random bytes: %v", err))
	}
	return rand.New(rand.NewSource(seed)).Float64
}

func run() error {
	filename := flag.Arg(0)
	src, err := readSource(filename)
	if err != nil {
		return err
	}

	if filename == "" || filename == "-" {
		filename = "<stdin>"
	}

	vm := goscript.New()
	vm.SetRandSource(newRandSource())

	new(require.Registry).Enable(vm)
	console.Enable(vm)

	_ = vm.Set("load", func(call goscript.FunctionCall) goscript.Value {
		return load(vm, call)
	})

	_ = vm.Set("readFile", func(name string) (string, error) {
		b, err := os.ReadFile(name)
		if err != nil {
			return "", err
		}
		return string(b), nil
	})

	if *timelimit > 0 {
		time.AfterFunc(time.Duration(*timelimit)*time.Second, func() {
			vm.Interrupt("timeout")
		})
	}

	prg, err := goscript.Compile(filename, string(src), false)
	if err != nil {
		return err
	}
	_, err = vm.RunProgram(prg)
	return err
}

func main() {
	defer func() {
		if x := recover(); x != nil {
			debug.Stack()
			panic(x)
		}
	}()
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		_ = pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if err := run(); err != nil {
		switch err := err.(type) {
		case *goscript.Exception:
			fmt.Println(err.String())
		case *goscript.InterruptedError:
			fmt.Println(err.String())
		default:
			fmt.Println(err)
		}
		os.Exit(64)
	}
}
