package lora_orm

import (
	"reflect"
	"strings"
)

/*
*@Author: LorraineWen
*支持嵌入式sql
 */
//嵌入式sql执行函数
func (session *Session) Exec(sql string, values ...any) (int64, error) {
	stmt, err := session.db.db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	r, err := stmt.Exec(values)
	if err != nil {
		return 0, err
	}
	if strings.Contains(strings.ToLower(sql), "insert") {
		return r.LastInsertId()
	}
	return r.RowsAffected()
}

// 单条语句查询
func (session *Session) QueryRow(sql string, data any, queryValues ...any) error {
	t := reflect.TypeOf(data)
	stmt, err := session.db.db.Prepare(sql)
	if err != nil {
		return err
	}
	rows, err := stmt.Query(queryValues...)
	if err != nil {
		return err
	}
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]any, len(columns))
	var fieldsScan = make([]any, len(columns))
	for i := range fieldsScan {
		fieldsScan[i] = &values[i]
	}
	if rows.Next() {
		err = rows.Scan(fieldsScan...)
		if err != nil {
			return err
		}
		v := reflect.ValueOf(data)
		valueOf := reflect.ValueOf(values)
		for i := 0; i < t.Elem().NumField(); i++ {
			name := t.Elem().Field(i).Name
			tag := t.Elem().Field(i).Tag
			sqlTag := tag.Get("lora_orm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(getTableFiledName(name))
			} else {
				if strings.Contains(sqlTag, ",") {
					sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
				}
			}
			for j, coName := range columns {
				if sqlTag == coName {
					if v.Elem().Field(i).CanSet() {
						covertValue := session.ConvertType(valueOf, j, v, i)
						v.Elem().Field(i).Set(covertValue)
					}
				}
			}
		}
	}
	return nil
}
