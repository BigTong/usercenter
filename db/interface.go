package db

import (
	"usercenter/user"
)

type UserDao interface {
	LoadUserList() ([]*user.User, error)
	AddUser([]*user.User) error

	UpdateUserRelations([]*user.UserRelationShip) error
	UpdateUserRelation(*user.UserRelationShip) (*user.UserRelationShip, error)
	GetUserRelation(int64) ([]*user.UserRelationShip, error)

	Close() error
}
