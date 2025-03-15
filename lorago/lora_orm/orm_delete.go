package lora_orm

import (
	"fmt"
	"strings"
)

func (session *Session) Delete() (int64, error) {
	query := fmt.Sprintf("delete from %s ", session.TableName)
	var sb strings.Builder
	sb.WriteString(query)
	sb.WriteString(session.whereParam.String())
	fmt.Println(sb.String())
	stmt, err := session.db.db.Prepare(sb.String())
	if err != nil {
		return 0, err
	}
	result, err := stmt.Exec(session.whereValues...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
