package lora_orm

/*
*@Author: LorraineWen
*支持单行插入和多行插入
 */
import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// 插入数据，实质上就是组装sql语句，调用的底层函数还是db.Prepare、Exec、LastInsertId
// 返回最后一行插入数据的id、插入的行数、错误
func (session *Session) Insert(data any) (int64, int64, error) {
	session.getFiledNames(data)
	query := fmt.Sprintf("insert into %s (%s) values(%s)", session.TableName, strings.Join(session.fieldName, ","), strings.Join(session.placeHolder, ","))
	stmt, err := session.db.db.Prepare(query)
	if err != nil {
		session.db.logger.Error(err)
		return -1, -1, err
	}
	r, err := stmt.Exec(session.values...)
	if err != nil {
		session.db.logger.Error(err)
		return -1, -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		session.db.logger.Error(err)
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		session.db.logger.Error(err)
		return -1, -1, err
	}
	return id, affected, nil
}

// 批量插入insert into user (user_name,password) values ("amie","123"),("miemie","234")
func (session *Session) BatchInsert(data []any) (int64, int64, error) {
	if len(data) == 0 {
		return -1, -1, errors.New("no data insert")
	}
	session.getBatchFieldNames(data)
	query := fmt.Sprintf("insert into %s (%s) values ", session.TableName, strings.Join(session.fieldName, ","))
	var sb strings.Builder
	sb.WriteString(query)
	for index, _ := range data {
		sb.WriteString("(")
		sb.WriteString(strings.Join(session.placeHolder, ","))
		sb.WriteString(")")
		if index < len(data)-1 {
			sb.WriteString(",")
		}
	}
	stmt, err := session.db.db.Prepare(sb.String())
	if err != nil {
		return -1, -1, err
	}
	r, err := stmt.Exec(session.values...)
	if err != nil {
		session.db.logger.Error(err)
		return -1, -1, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		session.db.logger.Error(err)
		return -1, -1, err
	}
	affected, err := r.RowsAffected()
	if err != nil {
		session.db.logger.Error(err)
		return -1, -1, err
	}
	return id, affected, nil
}

// 获取结构体的属性名称，支持自动将属性名称，映射为表字段名
func (session *Session) getFiledNames(data any) {
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	//如果传入的不是对象指针就报错
	if t.Kind() != reflect.Ptr {
		panic("data must be a pointer")
	}
	typeElem := t.Elem()
	valueElem := v.Elem()
	//如果没有传入表名，那么就通过prefix和结构体名拼接表名
	if session.TableName == "" {
		session.TableName = session.db.Prefix + strings.ToLower(typeElem.Name())
	}
	var fieldNames []string
	var placeholder []string
	var values []any
	for i := 0; i < typeElem.NumField(); i++ {
		//CanInterface会检测Value是否是可导出的，比如首字母大写的结构体属性或者函数
		//属性名开头是小写的不进行处理
		if !valueElem.Field(i).CanInterface() {
			continue
		}
		//解析tag
		field := typeElem.Field(i)
		sqlTag := field.Tag.Get("lora_orm")
		if sqlTag == "" {
			//如果用户没有通过lorasql标签指定属性为对应表名，UserName `lorasql:user_name`，就自动进行处理
			sqlTag = getTableFiledName(field.Name)
		}
		//id自增处理
		contains := strings.Contains(sqlTag, "auto_increment")
		if sqlTag == "id" || contains {
			//对id做个判断 如果其值小于等于0 数据库可能是自增 跳过此字段
			if isAutoId(valueElem.Field(i).Interface()) {
				continue
			}
		}
		if contains {
			sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
		}
		fieldNames = append(fieldNames, sqlTag)
		placeholder = append(placeholder, "?")
		values = append(values, valueElem.Field(i).Interface())
	}
	session.fieldName = fieldNames
	session.placeHolder = placeholder
	session.values = values
}
func isAutoId(id any) bool {
	t := reflect.TypeOf(id)
	v := reflect.ValueOf(id)
	switch t.Kind() {
	case reflect.Int64:
		if v.Interface().(int64) <= 0 {
			return true
		}
	case reflect.Int32:
		if v.Interface().(int32) <= 0 {
			return true
		}
	case reflect.Int:
		if v.Interface().(int) <= 0 {
			return true
		}
	default:
		return false
	}
	return false
}

// 将属性名转换为表字段名称，UserName变为user_name
func getTableFiledName(name string) string {
	all := name[:]
	var sb strings.Builder
	lastIndex := 0
	for index, value := range all {
		if value >= 65 && value <= 90 {
			if index == 0 {
				continue
			}
			sb.WriteString(name[lastIndex:index])
			sb.WriteString("_")
			lastIndex = index
		}
	}
	if lastIndex != len(name)-1 {
		sb.WriteString(name[lastIndex:])
	}
	return strings.ToLower(sb.String())
}

// 批量获取属性名和对应的值
func (session *Session) getBatchFieldNames(dataArray []any) {
	data := dataArray[0]
	t := reflect.TypeOf(data)
	v := reflect.ValueOf(data)
	if t.Kind() != reflect.Pointer {
		panic(errors.New("batch insert element type must be pointer"))
	}
	typeElem := t.Elem()
	valueElem := v.Elem()
	if session.TableName == "" {
		session.TableName = session.db.Prefix + strings.ToLower(getTableFiledName(typeElem.Name()))
	}
	var fieldNames []string
	var placeholder []string
	for i := 0; i < typeElem.NumField(); i++ {
		if !valueElem.Field(i).CanInterface() {
			continue
		}
		field := typeElem.Field(i)
		sqlTag := field.Tag.Get("lora_orm")
		if sqlTag == "" {
			sqlTag = strings.ToLower(getTableFiledName(field.Name))
		}
		contains := strings.Contains(sqlTag, "auto_increment")
		if sqlTag == "id" || contains {
			if isAutoId(valueElem.Field(i).Interface()) {
				continue
			}
		}
		if contains {
			sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
		}
		fieldNames = append(fieldNames, sqlTag)
		placeholder = append(placeholder, "?")
	}
	session.fieldName = fieldNames
	session.placeHolder = placeholder
	var allValues []any
	for _, value := range dataArray {
		t = reflect.TypeOf(value)
		v = reflect.ValueOf(value)
		typeElem = t.Elem()
		valueElem = v.Elem()
		for i := 0; i < typeElem.NumField(); i++ {
			if !valueElem.Field(i).CanInterface() {
				continue
			}
			field := typeElem.Field(i)
			sqlTag := field.Tag.Get("lora_orm")
			if sqlTag == "" {
				sqlTag = strings.ToLower(getTableFiledName(field.Name))
			}
			contains := strings.Contains(sqlTag, "auto_increment")
			if sqlTag == "id" || contains {
				if isAutoId(valueElem.Field(i).Interface()) {
					continue
				}
			}
			if contains {
				sqlTag = sqlTag[:strings.Index(sqlTag, ",")]
			}
			allValues = append(allValues, valueElem.Field(i).Interface())
		}
	}
	session.values = allValues
}
