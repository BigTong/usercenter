package usercenter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"usercenter/user"

	//"github.com/gorilla/mux"
	"github.com/BigTong/common/log"
	"github.com/juju/ratelimit"
	"github.com/zheng-ji/goSnowFlake"
)

const (
	SERVER_PATH = "/users"
)

type ErrResp struct {
	Status int    `json:"status"`
	Error  string `json:"error"`
}

func GetErrorMsg(errMsg string) string {
	errResp := &ErrResp{Status: 200, Error: errMsg}
	data, _ := json.Marshal(errResp)
	return string(data)
}

func NewUserServer() *UserServer {
	idGenerater, err := goSnowFlake.NewIdWorker(1)
	if err != nil {
		log.FFatal("new id generater get error:%s", err.Error())
	}

	return &UserServer{
		userDataCenter:          NewUserDataCenter(),
		userRelationCenter:      NewRelationShipCenter(),
		idGenerater:             idGenerater,
		userReadBucket:          ratelimit.NewBucketWithRate(50000, 100000),
		userWriteBucket:         ratelimit.NewBucketWithRate(30000, 60000),
		userRelationAddBucket:   ratelimit.NewBucketWithRate(20000, 40000),
		userRelationsReadBucket: ratelimit.NewBucketWithRate(30000, 60000),
	}
}

type UserServer struct {
	userDataCenter     *UserDataCenter
	userRelationCenter *RelationShipCenter
	idGenerater        *goSnowFlake.IdWorker

	userReadBucket  *ratelimit.Bucket
	userWriteBucket *ratelimit.Bucket

	userRelationAddBucket   *ratelimit.Bucket
	userRelationsReadBucket *ratelimit.Bucket
}

func (self *UserServer) ShutdownRacefully() bool {
	signal := make(chan int, 2)

	go func() {
		if self.userDataCenter.WaitingForDataWriteFinished() {
			signal <- 1
		}
	}()

	go func() {
		if self.userRelationCenter.waitingForDataWriteFinished() {
			signal <- 1
		}
	}()

	for {
		if len(signal) == 2 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	return true
}

func (self *UserServer) GetRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	hasToken := self.userRelationsReadBucket.WaitMaxDuration(1, 10*time.Millisecond)
	if !hasToken {
		self.writeErrorMessage("no token", w)
		return
	}
	if r.Method == "GET" {
		userId := StringToInt64(GetUrlPathArg(r.URL.Path, 2))
		if !user.CheckUsrIdValid(userId) {
			self.writeErrorMessage(fmt.Sprintf("not valid userId: %d", userId), w)
			return
		}
		userRelationShip := self.userRelationCenter.GetUserRelationShip(userId)
		log.FInfo("process one get user relationship req:%s", userRelationShip)
		fmt.Fprintf(w, "%s", userRelationShip)
		return
	}
	self.writeErrorMessage("not support method: "+r.Method, w)
}

type RelationShipPutData struct {
	State string `json:"state"`
}

func (self *UserServer) PutRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	hasToken := self.userRelationAddBucket.WaitMaxDuration(1, 10*time.Millisecond)
	if !hasToken {
		self.writeErrorMessage("no token", w)
		return
	}
	if r.Method == "PUT" {
		userId := StringToInt64(GetUrlPathArg(r.URL.Path, 2))
		otherUserId := StringToInt64(GetUrlPathArg(r.URL.Path, 4))
		if !user.CheckUsrIdValid(userId) ||
			!user.CheckUsrIdValid(otherUserId) {
			self.writeErrorMessage(
				fmt.Sprintf("not valied user id: %d, other userid: %d",
					userId, otherUserId), w)
			return
		}

		data, err := ReadHttpRequestBody(r)
		if err != nil {
			self.writeErrorMessage("read post body get error: "+err.Error(), w)
			return
		}

		decoder := json.NewDecoder(bytes.NewReader(data))
		relationPutData := &RelationShipPutData{}
		if err := decoder.Decode(relationPutData); err != nil {
			self.writeErrorMessage("decode put data get error: "+err.Error(), w)
			log.FInfo("body:%s", string(data))
			return
		}
		state := relationPutData.State
		if !user.CheckRelationStateValid(state) {
			self.writeErrorMessage(fmt.Sprintf("not valid state:%s", state), w)
			return
		}
		log.FInfo("userId: %d otherUserId: %d state:%s", userId, otherUserId, state)
		relation := user.NewUserRelation(userId, otherUserId, state)
		resp := self.userRelationCenter.UpdateRelationShip(relation)
		log.FInfo("process one update relationship req:%s", resp)
		fmt.Fprintf(w, "%s", resp)
		return
	}
	self.writeErrorMessage("not support method: "+r.Method, w)
}

func (self *UserServer) UserRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		self.showAllUsersHandler(w, r)
		return
	} else if r.Method == "POST" {
		self.userAddHandler(w, r)
		return
	}
	self.writeErrorMessage("not support method: "+r.Method, w)
}

type UserPostData struct {
	Name string `json:"name"`
}

func (self *UserServer) userAddHandler(w http.ResponseWriter, r *http.Request) {
	/*
		hasToken := self.userWriteBucket.WaitMaxDuration(1, 10*time.Millisecond)
		if !hasToken {
			self.writeErrorMessage("no token", w)
			return
		}
	*/

	startTime := time.Now().UnixNano()

	data, err := ReadHttpRequestBody(r)
	if err != nil {
		self.writeErrorMessage("read post body get error: "+err.Error(), w)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	userPostData := &UserPostData{}
	if err := decoder.Decode(userPostData); err != nil {
		self.writeErrorMessage("decode post data get error: "+err.Error(), w)
		log.FInfo("body:%s", string(data))
		return
	}

	endOne := time.Now().UnixNano()

	userName := userPostData.Name
	ok := self.userDataCenter.CheckNameRepeadedAndUpdateNameSet(userName)
	if !ok {
		self.writeErrorMessage(
			"not valid username: "+userName+", please change user name!", w)
		return
	}

	endTwo := time.Now().UnixNano()

	id, err := self.idGenerater.NextId()
	if err != nil {
		self.writeErrorMessage("id generater get error: "+err.Error(), w)
		return
	}

	endThree := time.Now().UnixNano()
	newUser := &user.User{
		Name:        userName,
		Id:          id,
		Createdtime: time.Now().UnixNano(),
	}
	resp, err := self.userDataCenter.AddUser(newUser)
	if err != nil {
		self.writeErrorMessage(" add user get error: "+err.Error(), w)
		return
	}
	log.FInfo("process one user add req: %s! get resp:",
		r.Host, string(resp))
	fmt.Fprintf(w, "%s", resp)
	endFour := time.Now().UnixNano()
	log.FInfo("cost time:%d %d %d %d",
		(endOne-startTime)/1000,
		(endTwo-endOne)/1000,
		(endThree-endTwo)/1000,
		(endFour-endThree)/1000)
}

func (self *UserServer) showAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	hasToken := self.userReadBucket.WaitMaxDuration(1, 10*time.Millisecond)
	if !hasToken {
		self.writeErrorMessage("no token", w)
		return
	}

	resp, err := self.userDataCenter.UserList()
	if err != nil {
		self.writeErrorMessage("some err with user list: "+err.Error(), w)
		return
	}
	log.FInfo("process one show user list req: %s ! get resp: %s",
		r.Host, string(resp))
	fmt.Fprintf(w, "%s", resp)
}

func (self *UserServer) writeErrorMessage(msg string, w http.ResponseWriter) {
	log.FInfo("%s", msg)
	fmt.Fprintf(w, "%s", GetErrorMsg(msg))
}
