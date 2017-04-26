package db

import (
	"usercenter/user"
)

type PostgresQlDb struct{}

func NewPostgresQlDb() *PostgresQlDb {
	return &PostgresQlDb{}
}

func (self *PostgresQlDb) AddUser(user []*user.User) (bool, error) {
	return true, nil
}

func (self *PostgresQlDb) LoadUserList() ([]*user.User, error) {
	return []*user.User{}, nil
}
