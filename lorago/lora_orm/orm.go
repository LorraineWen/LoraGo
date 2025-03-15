package lora_orm

/*
*@Author: LorraineWen
*支持orm操作
 */
import (
	"database/sql"
	"github.com/LorraineWen/lorago/lora_log"
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
	db          *Db
	TableName   string
	fieldName   []string //表的字段名
	placeHolder []string //字段占位符
	values      []any    //字段的值
}

func (session *Session) SetTableName(tableName string) *Session {
	session.TableName = tableName
	return session
}
