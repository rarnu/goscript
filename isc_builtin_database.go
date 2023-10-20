package goscript

import (
	"database/sql"
	"strings"
)

// migrate from gobase

func DatabaseQuery(db *sql.DB, sql string, args ...any) ([]map[string]string, error) {
	rows, err := db.Query(sql, args...)
	if err != nil {
		return nil, err
	}
	return fetchRows(rows, err)
}

func fetchRows(rows *sql.Rows, err error) ([]map[string]string, error) {
	if rows == nil || err != nil {
		return nil, err
	}

	fields, _ := rows.Columns()
	for k, v := range fields {
		fields[k] = camelCase(v)
	}
	columnsLength := len(fields)

	values := make([]string, columnsLength)
	args := make([]any, columnsLength)
	for i := 0; i < columnsLength; i++ {
		args[i] = &values[i]
	}

	index := 0
	listLength := 100
	lists := make([]map[string]string, listLength, listLength)
	for rows.Next() {
		if e := rows.Scan(args...); e == nil {
			row := make(map[string]string, columnsLength)
			for i, field := range fields {
				row[field] = values[i]
			}

			if index < listLength {
				lists[index] = row
			} else {
				lists = append(lists, row)
			}
			index++
		}
	}

	_ = rows.Close()

	return lists[0:index], nil
}

func camelCase(str string) string {
	if strings.Contains(str, "_") {
		items := strings.Split(str, "_")
		arr := make([]string, len(items))
		for k, v := range items {
			if 0 == k {
				arr[k] = v
			} else {
				arr[k] = strings.ToTitle(v)
			}
		}
		str = strings.Join(arr, "")
	}
	return str
}
