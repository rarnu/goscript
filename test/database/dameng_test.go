package database

import (
	"database/sql"
	"fmt"
	_ "gitee.com/chunanyong/dm"
	d0 "github.com/isyscore/isc-gobase/database"
	"testing"
	"time"
)

func TestDameng(t *testing.T) {

	connStr := fmt.Sprintf("dm://%s:%s@%s:%d", "SYSDBA", "SYSDBA001", "10.211.55.23", 5236)
	db, err := sql.Open("dm", connStr)
	db.SetConnMaxIdleTime(time.Second)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = db.Close()
	}()

	tx, _ := db.Begin()
	r, _ := tx.Exec("")
	r.LastInsertId()

	rows, _ := d0.Query(db, "select * from SYSDBA.SAMPLE")
	fmt.Printf("query\n")

	for idx, row := range rows {
		fmt.Printf("row: %d\n", idx)
		for k, v := range row {
			fmt.Printf("[%s] = %s\n", k, v)
		}
	}

}
