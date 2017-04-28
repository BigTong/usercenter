package usercenter

import (
	"encoding/json"
	"flag"
	"log"
	"sync"

	"usercenter/db"
	"usercenter/user"
)

const (
	DEFAULT_USER_LIST_Len   = 10000
	DEFAULT_NEW_USER_CHAN   = 100
	DEFAULT_BATCH_WRITE_NUM = 1
)

const (
	ERROR_NAME_REPEATED = "repeated user name"
)

var (
	postgresdbConfigFile = flag.String("postgres_config",
		"conf/postgresdb_config.json", "")
)

func NewUserDataCenter() *UserDataCenter {
	ret := &UserDataCenter{
		userList:       make([]*user.User, DEFAULT_USER_LIST_Len),
		userListRWLock: &sync.RWMutex{},

		userNameSet:   make(map[string]struct{}),
		userNameMutex: &sync.Mutex{},

		updateUserChan: make(chan *user.User, DEFAULT_NEW_USER_CHAN),

		postgresDb:        db.NewPostgresQlDb(*postgresdbConfigFile),
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

func (self *UserDataCenter) CheckNameRepeadedAndUpdateNameSet(userName string) bool {
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

	self.updateUserChan <- user
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
	users := make([]*user.User, DEFAULT_BATCH_WRITE_NUM)
	for {
		user := <-self.updateUserChan
		users = append(users, user)
		cnt++
		if (self.needFlushUserData && len(self.updateUserChan) == 0) ||
			cnt == DEFAULT_BATCH_WRITE_NUM {
			err := self.postgresDb.AddUser(users)
			if err != nil {
				panic("write db get error:" + err.Error())
			}
			cnt = 0
			users = users[:0]
		}

	}
}
