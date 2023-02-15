package test

import (
	c0 "context"
	"fmt"
	v2 "github.com/influxdata/influxdb-client-go/v2"
	// http2 "github.com/influxdata/influxdb-client-go/v2/api/http"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/console"
	"github.com/rarnu/goscript/module/require"
	"golang.org/x/net/context"
	"testing"
)

const _test_token = "CzUHhQBajU8O7mdLLj4IovzUy8eXf7rjCefqZlR-HwwoC0Tn71lAPJR4QqlzjVbDQ-wl6sZpcXXkD74g7ZLUSg=="
const _test_token2 = "rarnu:Rarnu1120"

func TestInfluxDB(t *testing.T) {

	cli := v2.NewClient("http://192.168.236.131:8086", _test_token)

	if cli == nil {
		return
	}
	defer cli.Close()
	wa := cli.WriteAPIBlocking("isc", "wms")

	/*
		wp := v2.NewPoint("stat", map[string]string{
			"unit": "temperature",
		}, map[string]any{
			"min": 30.0,
			"max": 35.0,
		}, time.Now())
		wa.WritePoint(wp)
	*/

	line := fmt.Sprintf("stat,unit=temperature avg=%f,max=%f", 23.5, 45.0)
	wa.WriteRecord(context.Background(), line)

	wa.Flush(context.Background())
}

func TestInfluxDBQuery(t *testing.T) {
	cli := v2.NewClient("http://192.168.236.131:8086", _test_token2)
	if cli == nil {
		return
	}
	defer cli.Close()
	qa := cli.QueryAPI("isc")
	query := `from(bucket:"wms")|> range(start: -1h) |> filter(fn: (r) => r._measurement == "stat")`
	result, _ := qa.Query(c0.Background(), query)
	defer func() {
		_ = result.Close()
	}()

	var _data []map[string]any

	for result.Next() {
		_data = append(_data, result.Record().Values())
		fmt.Printf("value: %+v\n", result.Record().Values())
	}

}

func TestInfluxDBScript(t *testing.T) {
	SCRIPT := `
	let db = new InfluxDB('http://10.211.55.23:8086','BnIerNUMg2Q5JJ7r7hfFwqF73owecYKoqR731vf202JXnJajGQLH7Q9Ti-S4X018QXRRU9WC_ZxzwNHDEndhrQ==')
	let wa = db.write('isyscore', 'wms')
	let pt = new InfluxDBPoint('stat', {'unit':'temperature'}, {'min': 25.0, 'max': 30.0})
	wa.writePoint(pt)
	db.close()
	`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	_, e := vm.RunString(SCRIPT)
	fmt.Printf("err: %v\n", e)
}
