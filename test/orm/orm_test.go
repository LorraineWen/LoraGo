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

func SaveUser() {
	dataSourceName := fmt.Sprintf("root:kylin.2023@tcp(localhost:3306)/lora_go?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
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
func TestInsert(t *testing.T) {
	SaveUser()
}
