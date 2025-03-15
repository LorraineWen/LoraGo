package orm

import (
	"fmt"
	"github.com/LorraineWen/lorago/lora_orm"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"testing"
)

type User struct {
	Id       int64
	UserName string
	Password string
	Age      int
}

func SaveUser(dataSourceName string) {
	db, err := lora_orm.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	user := &User{}
	user.UserName = "amie"
	user.Password = "123456"
	user.Age = 30
	id, aff, err := db.NewSession().SetTableName("user").Insert(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(id, aff, user)
}
func BatchInsertUser(dataSourceName string) {
	db, err := lora_orm.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	user := &User{}
	user.UserName = "mszlu"
	user.Password = "123456"
	user.Age = 30
	user1 := &User{}
	user1.UserName = "mszlu1"
	user1.Password = "1234567"
	user1.Age = 28
	var users []any
	users = append(users, user)
	users = append(users, user1)
	id, aff, err := db.NewSession().BatchInsert(users)
	if err != nil {
		panic(err)
	}
	fmt.Println(id, aff, users)
}
func TestInsert(t *testing.T) {
	dataSourceName := fmt.Sprintf("root:kylin.2023@tcp(localhost:3306)/lora_go?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	//SaveUser(dataSourceName)
	BatchInsertUser(dataSourceName)
}
