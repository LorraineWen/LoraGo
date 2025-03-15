package lora_orm

/*
*@Author: LorraineWen
*数据更新
 */
import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// 支持对单一字段的修改，Update("user_name","amie")，支持整个元组的修改Update(user)

func (s *Session) Update(data ...any) (int64, error) {
	//Update("age",1) or Update(user)
	size := len(data)
	if size <= 0 || size > 2 {
		return -1, errors.New("params error")
	}
	single := true
	if size == 2 {
		single = false
	}
	if !single {
		if s.updateParam.String() != "" {
			s.updateParam.WriteString(",")
		}
		field := data[0].(string)
		s.updateParam.WriteString(field)
		s.updateParam.WriteString(" = ?")
		s.updateValues = append(s.updateValues, data[1])
	} else {
		d := data[0]
		t := reflect.TypeOf(d)
		v := reflect.ValueOf(d)
		if t.Kind() != reflect.Pointer {
			return -1, errors.New("data not pointer")
		}
		tVar := t.Elem()
		vVar := v.Elem()
		if s.TableName == "" {
			s.TableName = s.db.Prefix + strings.ToLower(getTableFiledName(tVar.Name()))
		}
		for i := 0; i < tVar.NumField(); i++ {
			if s.updateParam.String() != "" {
				s.updateParam.WriteString(",")
			}
			sqlTag := tVar.Field(i).Tag.Get("mssql")
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
			s.updateParam.WriteString(sqlTag)
			s.updateParam.WriteString(" = ?")
			s.updateValues = append(s.updateValues, fieldValue)
		}
	}
	query := fmt.Sprintf("update %s set %s %s", s.TableName, s.updateParam.String(), s.whereParam.String())
	stmt, err := s.db.db.Prepare(query)
	if err != nil {
		return -1, err
	}
	s.updateValues = append(s.updateValues, s.values...)
	r, err := stmt.Exec(s.updateValues...)
	if err != nil {
		return -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		return -1, err
	}
	return affected, nil
}

func (s *Session) Where(field string, data any) *Session {
	if s.whereParam.String() != "" {
		s.whereParam.WriteString(" and ")
	} else {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" = ?")
	s.values = append(s.values, data)
	return s
}

func (s *Session) Or(field string, data any) *Session {
	if s.whereParam.String() != "" {
		s.whereParam.WriteString(" or ")
	} else {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" = ?")
	s.values = append(s.values, data)
	return s
}
