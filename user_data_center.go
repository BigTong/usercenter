package usercenter

import (
	"encoding/json"
	"log"
	"sync"

	"usercenter/db"
	"usercenter/user"
)

const (
	DefaultUserListLen     = 10000
	DefaultNewUserChan     = 100
	DefaultBatchWriteDbNum = 10
)

const (
	ERROR_NAME_REPEATED = "repeated user name"
)

func NewUserDataCenter() *UserDataCenter {
	ret := &UserDataCenter{
		userList:       make([]*user.User, DefaultUserListLen),
		userListRWLock: &sync.RWMutex{},

		userNameSet:   make(map[string]struct{}),
		userNameMutex: &sync.Mutex{},

		updateUserChan: make(chan *user.User, DefaultNewUserChan),

		postgresDb:        db.NewPostgresQlDb(),
		needFlushUserData: false,
	}

	ret.loadUserListFromDb()
	go ret.WriteNewUserDataToDb()
	return ret
}

type UserDataCenter struct {
	userList       []*user.User
	userListRWLock *sync.RWMutex

	// dedupe for user
	userNameSet   map[string]struct{}
	userNameMutex *sync.Mutex

	// user chan to write db
	updateUserChan chan *user.User

	postgresDb        db.UserDao
	needFlushUserData bool
}

func (self *UserDataCenter) CheckValidAndUpdateForUser(userName string) bool {
	if len(userName) == 0 {
		return false
	}

	self.userNameMutex.Lock()
	defer self.userNameMutex.Unlock()
	if _, ok := self.userNameSet[userName]; ok {
		return false
	}

	self.userNameSet[userName] = struct{}{}
	return true

}

func (self *UserDataCenter) AddUser(user *user.User) ([]byte, error) {
	self.userListRWLock.Lock()
	self.userList = append(self.userList, user)
	self.userListRWLock.Unlock()

	return json.Marshal(user)
}

func (self *UserDataCenter) UserList() ([]byte, error) {
	self.userListRWLock.RLock()
	defer self.userListRWLock.RUnlock()
	return json.Marshal(self.userList)
}

func (self *UserDataCenter) WriteUserDataFinished() bool {
	self.needFlushUserData = true
	return len(self.updateUserChan) == 0
}

func (self *UserDataCenter) loadUserListFromDb() {
	var err error
	self.userList, err = self.postgresDb.LoadUserList()
	if err != nil {
		log.Fatalln("load user list get error: ", err.Error())
	}

	for _, u := range self.userList {
		self.userNameSet[u.Name] = struct{}{}
	}

}

func (self *UserDataCenter) WriteNewUserDataToDb() {
	cnt := 0
	users := make([]*user.User, DefaultBatchWriteDbNum)
	for {
		user := <-self.updateUserChan
		users = append(users, user)
		cnt++
		if (self.needFlushUserData && len(self.updateUserChan) == 0) ||
			cnt == DefaultBatchWriteDbNum {
			self.postgresDb.AddUser(users)
			cnt = 0
			users = users[:0]
		}

	}
}
