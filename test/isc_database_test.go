package test

import (
	"database/sql"
	"fmt"
	d0 "github.com/isyscore/isc-gobase/database"
	"testing"
)

func testWithDatabase(connstr string, typ string, sqlstr string) {
	db, err := sql.Open(typ, connstr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = db.Close()
	}()
	rows, _ := d0.Query(db, sqlstr)
	for idx, row := range rows {
		fmt.Printf("row: %d\n", idx)
		for k, v := range row {
			fmt.Printf("[%s] = %s\n", k, v)
		}
	}
}

func TestDameng(t *testing.T) {
	connStr := fmt.Sprintf("dm://%s:%s@%s:%d", "SYSDBA", "SYSDBA001", "10.211.55.23", 5236)
	sqlstr := "select * from SYSDBA.SAMPLE"
	testWithDatabase(connStr, "dm", sqlstr)
}

func TestOracle(t *testing.T) {
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s", "system", "system", "10.211.55.23", 1521, "helowin")
	sqlstr := "select * from SYSTEM.SAMPLE"
	testWithDatabase(connStr, "oracle", sqlstr)
}

func TestSQLServer(t *testing.T) {
	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s", "sa", "@cw424.OMI", "10.211.55.23", 1433, "master")
	sqlstr := "select * from dbo.SAMPLE"
	testWithDatabase(connStr, "sqlserver", sqlstr)
}
