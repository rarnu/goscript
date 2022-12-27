package database

import (
	"database/sql"
	"fmt"
	d0 "github.com/isyscore/isc-gobase/database"
	_ "github.com/sijms/go-ora/v2"
	"testing"
)

func TestOracle(t *testing.T) {
	connStr := fmt.Sprintf("oracle://%s:%s@%s:%d/%s", "system", "system", "10.211.55.23", 1521, "helowin")
	db, err := sql.Open("oracle", connStr)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = db.Close()
	}()
	rows, _ := d0.Query(db, "select * from SYSTEM.SAMPLE")
	fmt.Printf("query\n")

	for idx, row := range rows {
		fmt.Printf("row: %d\n", idx)
		for k, v := range row {
			fmt.Printf("[%s] = %s\n", k, v)
		}
	}
}
