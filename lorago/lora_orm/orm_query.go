package lora_orm

/*
*@Author: LorraineWen
*支持查询后直接赋值给结构体对象，支持多行查询，支持只返回指定的字段值
 */
import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// 单行查询
func (session *Session) SelectOne(data any, fields ...string) error {
	t := reflect.TypeOf(data)
	var fieldStr = "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data type must be pointer"))
	}
	query := fmt.Sprintf("select %s from %s ", fieldStr, session.TableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(session.whereParam.String())
	session.db.logger.Info(sb.String())
	stmt, err := session.db.db.Prepare(sb.String())
	if err != nil {
		return err
	}
	rows, err := stmt.Query(session.values...)
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
	//映射表的字段到结构体的属性名
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

func (session *Session) ConvertType(valueOf reflect.Value, j int, v reflect.Value, i int) reflect.Value {
	valueElem := valueOf.Index(j)
	t2 := v.Elem().Field(i).Type()
	of := reflect.ValueOf(valueElem.Interface())
	covertValue := of.Convert(t2)
	return covertValue
}

// 多行查询
func (session *Session) Select(data any, fields ...string) ([]any, error) {
	var fieldStr = "*"
	if len(fields) > 0 {
		fieldStr = strings.Join(fields, ",")
	}
	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("data type must be struct"))
	}
	if session.TableName == "" {
		session.TableName = session.db.Prefix + strings.ToLower(getTableFiledName(t.Elem().Name()))
	}
	query := fmt.Sprintf("select %s from %s ", fieldStr, session.TableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(session.whereParam.String())
	stmt, err := session.db.db.Prepare(sb.String())
	if err != nil {
		return nil, err
	}
	rows, err := stmt.Query(session.whereValues...)
	if err != nil {
		return nil, err
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	values := make([]any, len(columns))
	var fieldsScan = make([]any, len(columns))
	for i := range fieldsScan {
		fieldsScan[i] = &values[i]
	}
	var results []any
	for {
		if rows.Next() {
			data = reflect.New(t.Elem()).Interface()
			err = rows.Scan(fieldsScan...)
			if err != nil {
				return nil, err
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
							eVar := valueOf.Index(j)
							t2 := v.Elem().Field(i).Type()
							of := reflect.ValueOf(eVar.Interface())
							covertValue := of.Convert(t2)
							v.Elem().Field(i).Set(covertValue)
						}
					}
				}
			}
			results = append(results, data)
		} else {
			break
		}
	}
	return results, nil
}

// 模糊查询
func (s *Session) Like(field string, data any) *Session {
	if s.whereParam.String() == "" {
		s.whereParam.WriteString(" where ")
	}
	s.whereParam.WriteString(field)
	s.whereParam.WriteString(" like ?")

	s.values = append(s.values, "%"+data.(string)+"%")
	return s
}

// 分组查询
func (s *Session) Group(field ...string) *Session {
	s.whereParam.WriteString(" group by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	return s
}

// 降序查询
func (s *Session) OrderDesc(field ...string) *Session {
	s.whereParam.WriteString(" order by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	s.whereParam.WriteString(" desc ")
	return s
}

// 升序查询
func (s *Session) OrderAsc(field ...string) *Session {
	s.whereParam.WriteString(" order by ")
	s.whereParam.WriteString(strings.Join(field, ","))
	s.whereParam.WriteString(" asc ")
	return s
}

// order by score asc,age desc，按照分数升序和年龄降序两种条件查询
func (s *Session) Order(field ...string) *Session {
	s.whereParam.WriteString(" order by ")
	size := len(field)
	if size%2 != 0 {
		panic("Order field must be even")
	}
	for index, v := range field {
		s.whereParam.WriteString(" ")
		s.whereParam.WriteString(v)
		s.whereParam.WriteString(" ")
		if index%2 != 0 && index < len(field)-1 {
			s.whereParam.WriteString(",")
		}
	}
	return s
}
