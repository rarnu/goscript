package goscript

import (
	"fmt"
	"testing"
)

func TestGoscript(t *testing.T) {
	script := NewGo(`
func goscript(ex, name string) (string, error) {
	greeting := ex + " " + name
	return greeting, nil
}
`)
	defer func(script *Script) {
		_ = script.Close()
	}(script)
	greeting, _ := script.Execute("Hello", "Rarnu")
	if greeting != "Hello Rarnu" {
		t.Fatalf("Expected greeting")
	}
}

type goScriptTest struct {
	script string
	inArgs []any
	outVal any
	outErr string
}

var tests = []goScriptTest{
	{script: `func goscript() (string, error) {
	return "Hello", nil
}`, outVal: "Hello"},
	{script: `func goscript(ex, name string) (string, error) {
	greeting := ex + " " + name
	return greeting, nil
}`, inArgs: []any{"Hello", "Rarnu"}, outVal: "Hello Rarnu"},
	{script: `import "strings"

func goscript(items ...string) (string, error) {
	return strings.Join(items, ","), nil
}`, inArgs: []any{"one", "two", "three"}, outVal: "one,two,three"},
	{script: `import "strings"

func goscript(separator string, items ...string) (string, error) {
	return strings.Join(items, separator), nil
}`, inArgs: []any{"|", "one", "two", "three"}, outVal: "one|two|three"},
	{script: `import "strings"

func goscript(separator string, items) (string, error) {
	return strings.Join(items, separator), nil
}`, inArgs: []any{"|", "one", "two", "three"}, outErr: "goscript:3:33: syntax error: mixed named and unnamed parameters"},
}

func TestGoscriptTests(t *testing.T) {
	for i := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			test := tests[i]
			script := NewGo(test.script)
			defer func(script *Script) {
				_ = script.Close()
			}(script)
			val, err := script.Execute(test.inArgs...)
			if err != nil {
				if err.Error() != test.outErr {
					t.Fatalf("error: executing with error: %v", err)
				}
			} else {
				if test.outErr != "" {
					t.Fatalf("error: executing fatal")
				}
			}
			if test.outVal != nil {
				if val != test.outVal {
					t.Fatalf("error: value mismatch")
				}
			}
		})
	}
}
