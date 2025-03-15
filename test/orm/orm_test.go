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
func Update(dataSourceName string) {
	db, err := lora_orm.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	user := &User{}
	user.UserName = "mszlu11111111111"
	user.Password = "123456111"
	user.Age = 3011
	update, err := db.NewSession().SetTableName("user").Where("id", 1).And().Where("age", 44).Update(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(update)
}
func QueryUser(dataSourceName string) {
	db, err := lora_orm.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	user := &User{}
	err = db.NewSession().SetTableName("user").Where("id", 1).SelectOne(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(user)
}
func QueryUsers(dataSourceName string) {
	db, err := lora_orm.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	user := &User{}
	users, err := db.NewSession().SetTableName("user").Where("age", 30).OrderDesc("age").Select(user)
	if err != nil {
		panic(err)
	}
	for _, u := range users {
		fmt.Println(u.(*User))
	}
}
func DeleteUser(dataSourceName string) {
	db, err := lora_orm.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	_, err = db.NewSession().SetTableName("user").Where("id", 1).Delete()
	if err != nil {
		panic(err)
	}
}
func GetTableTotalCount(dataSourceName string) {
	db, err := lora_orm.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	count, err := db.NewSession().SetTableName("user").TotalCount()
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
}
func TestInsert(t *testing.T) {
	dataSourceName := fmt.Sprintf("root:kylin.2023@tcp(localhost:3306)/lora_go?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))
	//SaveUser(dataSourceName)
	//BatchInsertUser(dataSourceName)
	//QueryUser(dataSourceName)
	//DeleteUser(dataSourceName)
	//QueryUsers(dataSourceName)
	GetTableTotalCount(dataSourceName)
}
