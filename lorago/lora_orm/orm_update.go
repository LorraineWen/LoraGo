package lora_orm

/*
*@Author: LorraineWen
*数据更新，支持Where关键字，支持单个字段和整个元组的更新，支持Or关键字
 */
import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// 支持对单一字段的修改，Update("user_name","amie")，支持整个元组的修改Update(user)
func (session *Session) Update(data ...any) (int64, error) {
	size := len(data)
	if size <= 0 || size > 2 {
		return -1, errors.New("params error")
	}
	single := true
	if size == 2 {
		single = false
	}
	if !single {
		if session.updateParam.String() != "" {
			session.updateParam.WriteString(",")
		}
		field := data[0].(string)
		session.updateParam.WriteString(field)
		session.updateParam.WriteString(" = ?")
		session.updateValues = append(session.updateValues, data[1])
	} else {
		d := data[0]
		t := reflect.TypeOf(d)
		v := reflect.ValueOf(d)
		if t.Kind() != reflect.Pointer {
			return -1, errors.New("data not pointer")
		}
		tVar := t.Elem()
		vVar := v.Elem()
		if session.TableName == "" {
			session.TableName = session.db.Prefix + strings.ToLower(getTableFiledName(tVar.Name()))
		}
		for i := 0; i < tVar.NumField(); i++ {
			if session.updateParam.String() != "" {
				session.updateParam.WriteString(",")
			}
			sqlTag := tVar.Field(i).Tag.Get("lora_orm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(getTableFiledName(tVar.Field(i).Name))
			}
			if strings.Contains(sqlTag, ",") {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
			fieldValue := vVar.Field(i).Interface()
			if sqlTag == "id" && isAutoId(fieldValue) {
				continue
			}
			session.updateParam.WriteString(sqlTag)
			session.updateParam.WriteString(" = ?")
			session.updateValues = append(session.updateValues, fieldValue)
		}
	}
	query := fmt.Sprintf("update %s set %s %s", session.TableName, session.updateParam.String(), session.whereParam.String())
	stmt, err := session.db.db.Prepare(query)
	if err != nil {
		return -1, err
	}
	session.updateValues = append(session.updateValues, session.values...)
	r, err := stmt.Exec(session.updateValues...)
	if err != nil {
		return -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, err
	}
	return affected, nil
}
