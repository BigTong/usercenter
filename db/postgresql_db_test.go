package db

import (
	"testing"
	"time"

	"usercenter/user"
)

var config = &PostgresDBConfig{
	User:     "postgres",
	Password: "zex",
	Database: "userdb",
}

var db = NewPostgredQlDbWithConfig(config)

func TestAddUser(t *testing.T) {
	newUser := &user.User{
		Name:        "Alice",
		Id:          1,
		Gender:      "m",
		Age:         18,
		Createdtime: time.Now().Unix()}
	users := []*user.User{}
	users = append(users, newUser)
	err := db.AddUser(users)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestLoadUserList(t *testing.T) {
	userList, err := db.LoadUserList()
	if err != nil {
		t.Error(err.Error())
	}
	t.Logf("len:%d", userList[0], len(userList))
	if len(userList) > 0 {
		t.Logf("first user: %s %d %s",
			userList[0].Name,
			userList[0].Createdtime,
			userList[0].Gender)
	}

}

func TestGetUserRelation(t *testing.T) {

}
