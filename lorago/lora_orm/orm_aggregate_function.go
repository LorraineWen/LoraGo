package lora_orm

import (
	"fmt"
	"strings"
)

/*
*@Author: LorraineWen
*支持聚合函数
 */
//计数聚合函数
func (session *Session) TotalCount() (int64, error) {
	return session.Aggregate("count", "*")
}

// 提供自定义聚合函数的统一实现
func (session *Session) Aggregate(funcName, field string) (int64, error) {
	var aggSb strings.Builder
	aggSb.WriteString(funcName)
	aggSb.WriteString("(")
	aggSb.WriteString(field)
	aggSb.WriteString(")")
	query := fmt.Sprintf("select %s from %s ", aggSb.String(), session.TableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(session.whereParam.String())
	session.db.logger.Info(sb.String())
	stmt, err := session.db.db.Prepare(sb.String())
	if err != nil {
		return 0, err
	}
	var result int64
	row := stmt.QueryRow()
	err = row.Err()
	if err != nil {
		return 0, err
	}
	err = row.Scan(&result)
	if err != nil {
		return 0, err
	}
	return result, nil
}
