package db

import (
	"usercenter/user"
)

type UserDao interface {
	LoadUserList() ([]*user.User, error)
	AddUser([]*user.User) (bool, error)
}
