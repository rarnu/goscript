package dap

import (
	"encoding/json"
	"fmt"
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

	initialize(t, c)

	launch(t, filePath, fileName, c)

	setBreakpoints(t, fileName, filePath, []int{2, 3, 5}, c)
	// stop at 【breakpoint line 2】
	onContinue(t, c)

	variables(t, c)
	// result = 1
	evaluate(t, "g_a = 1", c)
	// = step over
	//next(t, c)

	//stepIn(t, c)
}

func stepIn(t *testing.T, c *Client) {
	stepInRequest := &dap.StepInRequest{Arguments: dap.StepInArguments{ThreadId: 1}}
	stepInResponse, err := c.onStepIn(stepInRequest)
	print("stepInResponse", stepInResponse, err, t)
}

func next(t *testing.T, c *Client) {
	nextRequest := &dap.NextRequest{Arguments: dap.NextArguments{ThreadId: 1}}
	nextResponse, err := c.OnNext(nextRequest)
	print("nextResponse", nextResponse, err, t)
}

func evaluate(t *testing.T, expression string, c *Client) {
	evaluateRequest := &dap.EvaluateRequest{Arguments: dap.EvaluateArguments{Expression: expression}}
	evaluateResponse, err := c.Evaluate(evaluateRequest)
	print("evaluateResponse", evaluateResponse, err, t)
}

func variables(t *testing.T, c *Client) {
	variablesRequest := &dap.VariablesRequest{}
	variablesResponse, err := c.OnVariables(variablesRequest)
	print("variablesResponse", variablesResponse, err, t)
}

func launch(t *testing.T, filePath string, fileName string, c *Client) {
	args := LaunchConfig{
		DlvCwd:  filePath,
		Program: fileName,
		NoDebug: false} //true=直接启动
	b, _ := json.Marshal(args)
	launchRequest := &dap.LaunchRequest{Arguments: b}
	launchResponse, err := c.Launch(launchRequest)
	print("launchResponse", launchResponse, err, t)
}

func initialize(t *testing.T, c *Client) {
	initializeRequest := &dap.InitializeRequest{
		Arguments: dap.InitializeRequestArguments{
			PathFormat:      "path",
			LinesStartAt1:   true,
			ColumnsStartAt1: true,
		}}
	initializeResponse, err := c.Initialize(initializeRequest)
	print("initializeResponse", initializeResponse, err, t)
}

func setBreakpoints(t *testing.T, fileName string, filePath string, lines []int, c *Client) {
	setBreakPointsRequest := &dap.SetBreakpointsRequest{
		Arguments: dap.SetBreakpointsArguments{
			Source: dap.Source{
				Name: fileName,
				Path: filePath,
			},
			Lines: lines,
		},
	}
	var bps []dap.SourceBreakpoint
	for _, l := range lines {
		bps = append(bps, dap.SourceBreakpoint{Line: l})
	}
	setBreakPointsRequest.Arguments.Breakpoints = bps

	setBreakpointsResponse, err := c.SetBreakpoints(setBreakPointsRequest)
	print("setBreakpointsResponse", setBreakpointsResponse, err, t)
}

func onContinue(t *testing.T, c *Client) {
	continueRequest := &dap.ContinueRequest{}
	continueResponse, err := c.OnContinue(continueRequest)
	print("continueResponse", continueResponse, err, t)
}

func print(prefix string, data any, err error, t *testing.T) {
	if err != nil {
		fmt.Println(prefix + ": " + err.Error())
	} else {
		if b, err := json.Marshal(data); err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println("【" + prefix + "】: " + string(b))
		}
	}
}
