package test

import (
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/console"
	"github.com/rarnu/goscript/module/require"
	"testing"
)

func TestHttpReq(t *testing.T) {
	SCRIPT := `
	let date = new Date();
let year = date.getFullYear();
let month = ("0" + (date.getMonth() + 1)).slice(-2); // 月份是从0开始的，所以需要+1
let day = ("0" + date.getDate()).slice(-2);

let yearStr = ''+year;
let formattedDate = yearStr+"-" + month+"-" + day;
let currDateStr  = formattedDate.toUpperCase()

let originalUrl = "http://10.0.108.11:34657/stage-api/";
let url = originalUrl + 'eam/EmInspectionTask/getPendingList';

let header = {
    "Authorization":"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE2OTkyNTY4MTAsInVzZXJfdHlwZSI6IjAwIiwidXNlcl9pZCI6LTIsInVzZXJfbmFtZSI6ImFkbWluIiwidXNlcl9rZXkiOiI3M2IyMzhjYy00YzVmLTQ1ZmEtYWFmZi1jZTI5ZWUwMTk4MWQiLCJlbnRlcnByaXNlX2lkIjotMSwidG9rZW5fdHlwZSI6ImFwaV90b2tlbnM6IiwiZW50ZXJwcmlzZV9uYW1lIjoiYWRtaW5pc3RyYXRvciIsImV4cCI6MjY0NTMzNjgxMCwibmJmIjoxNjk5MjU2ODEwfQ.YsoM0l49rr7KCGr4pr6D3heBPIJsVD57bij80892K9A"
};

let data = {
    "lineId": "1610173878284455936",
    "workDivision": "Inspection_Daily",
    "dateRange": [currDateStr, currDateStr],
    "inspectionTaskStatusList": ["Planning"],
    "pageNum": "1",
    "pageSize": "10"
}

let resp = HTTP.get(url, header, data);

resp;
	`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	v0, err := vm.RunString(SCRIPT)
	vx := v0.Export()
	t.Logf("v0 = %+v\n, err = %+v\n", vx, err)
}
