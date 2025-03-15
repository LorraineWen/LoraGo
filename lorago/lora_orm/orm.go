package lora_orm

/*
*@Author: LorraineWen
*支持orm操作
 */
import (
	"database/sql"
	"errors"
	"github.com/LorraineWen/lorago/lora_log"
	"reflect"
	"strings"
	"time"
)

type Db struct {
	db     *sql.DB
	logger *lora_log.Logger
	Prefix string //如果TableName为空，就会根据用户传入的Prefix、反射获取的结构体名称来组装一个表名，比如prefix="p4_"，结构体名字是User，得到p4_user表名
}

func Open(driver string, source string) (*Db, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}
	loraDb := &Db{
		db:     db,
		logger: lora_log.NewLogger(),
	}
	db.SetMaxIdleConns(5)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetConnMaxIdleTime(time.Minute * 1)
	//判断能否连接到该数据库
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return loraDb, nil
}

// 最大连接数
func (db *Db) SetMaxOpenConns(n int) {
	db.db.SetMaxOpenConns(n)
}

// 最大空闲连接数
func (db *Db) SetMaxIdleConns(n int) {
	db.db.SetMaxIdleConns(n)
}

// 连接最大存活时间
func (db *Db) SetConnMaxLifetime(d time.Duration) {
	db.db.SetConnMaxLifetime(d)
}

// 空闲连接最大存活时间
func (db *Db) SetConnMaxIdleTime(d time.Duration) {
	db.db.SetConnMaxIdleTime(d)
}
func (db *Db) SetTablePrefix(prefix string) *Db {
	db.Prefix = prefix
	return db
}

func (db *Db) NewSession() *Session {
	return &Session{db: db}
}

type Session struct {
	db           *Db
	TableName    string
	fieldName    []string        //表的字段名
	placeHolder  []string        //字段占位符
	values       []any           //字段的值
	updateValues []any           //update参数的值
	updateParam  strings.Builder //update参数
	whereParam   strings.Builder //where关键字的参数
}

func (session *Session) SetTableName(tableName string) *Session {
	session.TableName = tableName
	return session
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
