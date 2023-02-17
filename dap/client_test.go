package dap

import (
	"encoding/json"
	"github.com/google/go-dap"
	"testing"
)

func Test_newClient(t *testing.T) {
	fileName := "test.js"
	filePath := "D:\\code\\goscript\\dap\\script"

	c, err := newClient("0.0.0.0:6666")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	initializeRequest := &dap.InitializeRequest{
		Arguments: dap.InitializeRequestArguments{
			PathFormat:      "path",
			LinesStartAt1:   true,
			ColumnsStartAt1: true,
		}}
	initializeResponse, err := c.Initialize(initializeRequest)
	print("initializeResponse", initializeResponse, err, t)

	args := LaunchConfig{
		DlvCwd:  filePath,
		Program: fileName,
		NoDebug: false} //true=直接启动
	b, _ := json.Marshal(args)
	launchRequest := &dap.LaunchRequest{Arguments: b}
	launchResponse, err := c.Launch(launchRequest)
	print("launchResponse", launchResponse, err, t)

	setBreakPointsRequest := &dap.SetBreakpointsRequest{
		Arguments: dap.SetBreakpointsArguments{
			Source: dap.Source{
				Name: fileName,
				Path: filePath,
			},
			Breakpoints: []dap.SourceBreakpoint{{Line: 2}, {Line: 3}, {Line: 5}},
		},
	}
	setBreakpointsResponse, err := c.SetBreakpoints(setBreakPointsRequest)
	print("setBreakpointsResponse", setBreakpointsResponse, err, t)
	// stop at 【breakpoint line 2】
	continueRequest := &dap.ContinueRequest{}
	continueResponse, err := c.OnContinue(continueRequest)
	print("continueResponse", continueResponse, err, t)

	variablesRequest := &dap.VariablesRequest{}
	variablesResponse, err := c.OnVariables(variablesRequest)
	print("variablesResponse", variablesResponse, err, t)

	evaluateRequest := &dap.EvaluateRequest{Arguments: dap.EvaluateArguments{Expression: "g_a = 1"}}
	evaluateResponse, err := c.Evaluate(evaluateRequest)
	print("evaluateResponse", evaluateResponse, err, t)

	//nextRequest := &dap.NextRequest{Arguments: dap.NextArguments{ThreadId: 1}}
	//nextResponse, err := c.OnNext(nextRequest)
	//print("nextResponse", nextResponse, err, t)
	//
	//stepInRequest := &dap.StepInRequest{Arguments: dap.StepInArguments{ThreadId: 1}}
	//stepInResponse, err := c.onStepIn(stepInRequest)
	//print("stepInResponse", stepInResponse, err, t)
}

func print(prefix string, data any, err error, t *testing.T) {
	if err != nil {
		t.Log(prefix + ": " + err.Error())
	} else {
		if b, err := json.Marshal(data); err != nil {
			t.Error(err)
		} else {
			t.Log(prefix + ": " + string(b))
		}
	}
}
