package lora_orm

/*
*@Author: LorraineWen
*支持事务，开启事务，提交事务，回滚事务
 */
//开启事务，底层是db.Begin函数
func (session *Session) Begin() error {
	tx, err := session.db.db.Begin()
	if err != nil {
		return err
	}
	session.tx = tx
	session.txStatus = true
	return nil
}

func (session *Session) Commit() error {
	err := session.tx.Commit()
	if err != nil {
		return err
	}
	session.txStatus = false
	return nil
}

func (session *Session) Rollback() error {
	err := session.tx.Rollback()
	if err != nil {
		return err
	}
	session.txStatus = false
	return nil
}
