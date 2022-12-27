package database

import (
	"database/sql"
	"fmt"
	_ "github.com/denisenkom/go-mssqldb"
	d0 "github.com/isyscore/isc-gobase/database"
	"testing"
)

func TestSQLServer(t *testing.T) {
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", "sa", "@cw424.OMI", "10.211.55.23", 1433, "master")
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = db.Close()
	}()
	rows, _ := d0.Query(db, "select * from dbo.SAMPLE")
	fmt.Printf("query\n")

	for idx, row := range rows {
		fmt.Printf("row: %d\n", idx)
		for k, v := range row {
			fmt.Printf("[%s] = %s\n", k, v)
		}
	}
}
