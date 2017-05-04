package usercenter

import (
	"encoding/json"
	"flag"
	"sync"
	"time"

	"usercenter/db"
	"usercenter/user"

	"github.com/BigTong/common/log"
)

const (
	DEFAULT_USER_LIST_Len   = 1000000
	DEFAULT_NEW_USER_CHAN   = 10000
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

func (udc *UserDataCenter) CheckNameRepeadedAndUpdateNameSet(userName string) bool {
	if len(userName) == 0 {
		return false
	}

	udc.userNameMutex.Lock()
	defer udc.userNameMutex.Unlock()
	if _, ok := udc.userNameSet[userName]; ok {
		return false
	}

	udc.userNameSet[userName] = struct{}{}
	return true

}

func (udc *UserDataCenter) AddUser(user *user.User) ([]byte, error) {
	udc.userListRWLock.Lock()
	udc.userList = append(udc.userList, user)
	udc.userListRWLock.Unlock()

	udc.updateUserChan <- user
	return json.Marshal(user)
}

func (udc *UserDataCenter) UserList() ([]byte, error) {
	udc.userListRWLock.RLock()
	defer udc.userListRWLock.RUnlock()
	return json.Marshal(udc.userList)
}

func (udc *UserDataCenter) WaitingForDataWriteFinished() bool {
	udc.needFlushUserData = true
	for {
		if len(udc.updateUserChan) == 0 {
			udc.postgresDb.Close()
			return true
		}
		time.Sleep(100 * time.Millisecond)
	}
	return true
}

func (udc *UserDataCenter) loadUserListFromDb() {
	var err error
	udc.userList, err = udc.postgresDb.LoadUserList()
	if err != nil {
		log.FFatal("load user list get error: %s", err.Error())
	}

	for _, u := range udc.userList {
		udc.userNameSet[u.Name] = struct{}{}
	}

}

func (udc *UserDataCenter) WriteNewUserDataToDb() {
	cnt := 0
	users := []*user.User{}
	for {
		user := <-udc.updateUserChan
		users = append(users, user)
		cnt++
		if (udc.needFlushUserData && len(udc.updateUserChan) == 0) ||
			cnt == DEFAULT_BATCH_WRITE_NUM {
			log.FInfo("get user:%v len:%d", *users[0], len(users))
			err := udc.postgresDb.AddUser(users)
			if err != nil {
				log.FFatal("write db get error:%s", err.Error())
			}
			cnt = 0
			users = users[:0]
		}

	}
}
