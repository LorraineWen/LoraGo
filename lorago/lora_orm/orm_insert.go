package lora_orm

/*
*@Author: LorraineWen
*支持单行插入和多行插入
 */
import (
	"errors"
	"fmt"
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
