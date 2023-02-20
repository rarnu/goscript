package dap

import (
	"encoding/json"
	"github.com/google/go-dap"
	"testing"
)

const SERVER_ADDRESS = "0.0.0.0:6666"

const SCRIPT = `
var g_a = 1;
var g_b = "abc";
var g_c = null;
var g_d = false;
var g_e = [1, 2];
var g_f = {a: 'foo', b: 42, c: {}};

function f1() {
	let a = 1;
	let b = "abc";
	let c = null;
	let d = false;
	let e = [1, 2];
	let f = {a: 'foo', b: 42, c: {}};
	a++;
	return 1;
}
f1();
`

func Test_newClientWithScript(t *testing.T) {
	fileName := "test.js"

	c, err := newClient(SERVER_ADDRESS)
	if err != nil {
		t.Fatal(err)
	}
	c.msgChan = make(chan string)
	go func() {
		for {
			msg := <-c.msgChan
			t.Log(msg)
		}
	}()

	defer c.Close()

	initialize(t, c)

	launch(t, "", fileName, SCRIPT, c)

	setBreakpoints(t, fileName, "", []int{2, 3, 5}, c)
	// stop at 【breakpoint line 2】
	onContinue(t, c)
	// stop at 【breakpoint line 3】
	onContinue(t, c)
	// stop at 【breakpoint line 5】
	onContinue(t, c)

	variables(t, c)
	// result = true
	evaluate(t, "g_a == 1", c)
	// result = false
	evaluate(t, "g_a == 2", c)
	// = step over
	next(t, c)

	stepIn(t, c)

	onDisconnect(t, c)

	t.Log("end!")
}

func Test_newClientWithFilePath(t *testing.T) {
	fileName := "test.js"
	filePath := "D:\\code\\goscript\\dap\\script"

	c, err := newClient(SERVER_ADDRESS)
	if err != nil {
		t.Fatal(err)
	}
	c.msgChan = make(chan string)
	go func() {
		for {
			msg := <-c.msgChan
			t.Log(msg)
		}
	}()

	defer c.Close()

	initialize(t, c)

	launch(t, filePath, fileName, SCRIPT, c)

	setBreakpoints(t, fileName, filePath, []int{2, 3, 5}, c)
	// stop at 【breakpoint line 2】
	onContinue(t, c)
	// stop at 【breakpoint line 3】
	onContinue(t, c)
	// stop at 【breakpoint line 5】
	onContinue(t, c)

	variables(t, c)
	// result = true
	evaluate(t, "g_a == 1", c)
	// result = false
	evaluate(t, "g_a == 2", c)
	// = step over
	next(t, c)

	stepIn(t, c)

	onDisconnect(t, c)

	t.Log("end!")
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
	t.Log("expression: " + expression + ", result :" + evaluateResponse.Body.Result)
}

func variables(t *testing.T, c *Client) {
	variablesRequest := &dap.VariablesRequest{}
	variablesResponse, err := c.OnVariables(variablesRequest)
	print("variablesResponse", variablesResponse, err, t)

	//b, _ := json.Marshal(variablesResponse.Body.Variables)
	//t.Log("variables :" + string(b))
	t.Logf("variables size : %d", len(variablesResponse.Body.Variables))
}

func launch(t *testing.T, filePath string, fileName string, script string, c *Client) {
	args := LaunchConfig{
		FilePath: filePath,
		Program:  fileName,
		Script:   script,
		NoDebug:  false} //true=直接启动
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

func onDisconnect(t *testing.T, c *Client) {
	disconnectRequest := &dap.DisconnectRequest{}
	disconnectResponse, err := c.OnDisconnectRequest(disconnectRequest)
	print("disconnectResponse", disconnectResponse, err, t)
}

func print(prefix string, data any, err error, t *testing.T) {
	if err != nil {
		t.Log(prefix + ": " + err.Error())
	} else {
		if _, err := json.Marshal(data); err != nil {
			t.Log(err.Error())
		} else {
			t.Log("  【" + prefix + "】")
			//t.Log("  【" + prefix + " size】: " + strconv.Itoa(len(b)))
			//t.Log("  【" + prefix + "】: " + string(b))
		}
	}
}
