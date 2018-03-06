package zxgo

import (
	"database/sql"
)

func QueryData(db *sql.DB, sql string) (results map[int]map[string]string, err error) {

	sqlRows, err := db.Query(sql)
	if err != nil {
		return
	}

	results, err = getSqlRowsData(sqlRows)

	return
}

func getSqlRowsData(sqlRows *sql.Rows) (results map[int]map[string]string, err error) {

	columns, err := sqlRows.Columns() //查询出各列的字段名,将其读出来.
	if err != nil {
		return
	}

	values := make([][]byte, len(columns))     //每个数据行中,各个字段的值,将它们获取到byte里
	scans := make([]interface{}, len(columns)) //因为每次查询出来的列是不定长的,用len(column)定住当次查询的长度
	for i := range values {
		scans[i] = &values[i]
	}

	results = make(map[int]map[string]string)

	for i := 0; sqlRows.Next(); i++ {
		if err2 := sqlRows.Scan(scans...); err2 != nil {
			err = err2
			return
		}
		row := make(map[string]string)
		for offset, v := range values {
			key := columns[offset]
			val := string(v)
			row[key] = val
		}
		results[i] = row
	}

	return
}
