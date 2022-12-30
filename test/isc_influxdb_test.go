package test

import (
	c0 "context"
	"fmt"
	v2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/rarnu/goscript"
	"github.com/rarnu/goscript/module/console"
	"github.com/rarnu/goscript/module/require"
	"testing"
	"time"
)

func TestInfluxDB(t *testing.T) {

	cli := v2.NewClient("http://10.211.55.23:8086", "BnIerNUMg2Q5JJ7r7hfFwqF73owecYKoqR731vf202JXnJajGQLH7Q9Ti-S4X018QXRRU9WC_ZxzwNHDEndhrQ==")

	if cli == nil {
		return
	}
	defer cli.Close()
	wa := cli.WriteAPI("isyscore", "wms")

	wp := v2.NewPoint("stat", map[string]string{
		"unit": "temperature",
	}, map[string]any{
		"min": 30.0,
		"max": 35.0,
	}, time.Now())
	wa.WritePoint(wp)
	wa.Flush()
}

func TestInfluxDBQuery(t *testing.T) {
	cli := v2.NewClient("http://10.211.55.23:8086", "BnIerNUMg2Q5JJ7r7hfFwqF73owecYKoqR731vf202JXnJajGQLH7Q9Ti-S4X018QXRRU9WC_ZxzwNHDEndhrQ==")
	if cli == nil {
		return
	}
	defer cli.Close()
	qa := cli.QueryAPI("isyscore")
	query := `from(bucket:"wms")|> range(start: -3h) |> filter(fn: (r) => r._measurement == "stat")`
	result, _ := qa.Query(c0.Background(), query)
	for result.Next() {
		fmt.Printf("value: %+v\n", result.Record().Values())
	}
}

func TestInfluxDBScript(t *testing.T) {
	SCRIPT := `
	let db = new InfluxDB('http://10.211.55.23:8086','BnIerNUMg2Q5JJ7r7hfFwqF73owecYKoqR731vf202JXnJajGQLH7Q9Ti-S4X018QXRRU9WC_ZxzwNHDEndhrQ==')
	let wa = db.write('isyscore', 'wms')
	wa.writeRecord('23333')
	db.close()
	`
	vm := goscript.New()
	new(require.Registry).Enable(vm)
	console.Enable(vm)
	_, _ = vm.RunScript("index.js", SCRIPT)
}
