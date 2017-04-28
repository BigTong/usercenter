package usercenter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"time"

	"usercenter/user"

	//"github.com/gorilla/mux"
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
		log.Fatalf("new id generater get error:%s", err.Error())
	}

	return &UserServer{
		userDataCenter:     NewUserDataCenter(),
		userRelationCenter: NewRelationShipCenter(),
		idGenerater:        idGenerater,
	}
}

type UserServer struct {
	userDataCenter     *UserDataCenter
	userRelationCenter *RelationShipCenter
	idGenerater        *goSnowFlake.IdWorker
}

func (self *UserServer) GetRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			statckBuf := make([]byte, 64*1024)
			runtime.Stack(statckBuf, false)
			log.Fatalf("processor manager recover:%v, stack%s",
				r, string(statckBuf))
		}
	}()

	if r.Method == "GET" {
		userId := StringToInt64(GetUrlPathArg(r.URL.Path, 2))
		if !user.CheckUsrIdValid(userId) {
			self.writeErrorMessage(fmt.Sprintf("not valid userId: %d", userId), w)
			return
		}
		userRelationShip := self.userRelationCenter.GetUserRelationShip(userId)
		log.Printf("process one get user relationship req:%s", userRelationShip)
		fmt.Fprintf(w, "%s", userRelationShip)
		return
	}
	self.writeErrorMessage("not support method: "+r.Method, w)
}

type RelationShipPutData struct {
	State string `json:"state"`
}

func (self *UserServer) PutRelationshipHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			statckBuf := make([]byte, 64*1024)
			runtime.Stack(statckBuf, false)
			log.Fatalf("processor manager recover:%v, stack%s",
				r, string(statckBuf))
		}
	}()
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
			log.Println("body: " + string(data))
			return
		}
		state := relationPutData.State
		if !user.CheckRelationStateValid(state) {
			self.writeErrorMessage(fmt.Sprintf("not valid state:%s", state), w)
			return
		}
		log.Printf("userId: %d otherUserId: %d state:%s", userId, otherUserId, state)
		relation := user.NewUserRelation(userId, otherUserId, state)
		resp := self.userRelationCenter.UpdateRelationShip(relation)
		log.Printf("process one update relationship req:%s", resp)
		fmt.Fprintf(w, "%s", resp)
		return
	}
	self.writeErrorMessage("not support method: "+r.Method, w)
}

func (self *UserServer) UserRequestHandler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			statckBuf := make([]byte, 64*1024)
			runtime.Stack(statckBuf, false)
			log.Fatalf("processor manager recover:%v, stack%s",
				r, string(statckBuf))
		}
	}()

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
	data, err := ReadHttpRequestBody(r)
	if err != nil {
		self.writeErrorMessage("read post body get error: "+err.Error(), w)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(data))
	userPostData := &UserPostData{}
	if err := decoder.Decode(userPostData); err != nil {
		self.writeErrorMessage("decode post data get error: "+err.Error(), w)
		log.Println("body: " + string(data))
		return
	}

	userName := userPostData.Name
	ok := self.userDataCenter.CheckNameRepeadedAndUpdateNameSet(userName)
	if !ok {
		self.writeErrorMessage(
			"not valid username: "+userName+", please change user name!", w)
		return
	}

	id, err := self.idGenerater.NextId()
	if err != nil {
		self.writeErrorMessage("id generater get error: "+err.Error(), w)
		return
	}

	newUser := &user.User{
		Name:        userName,
		Id:          id,
		Createdtime: time.Now().Unix(),
	}
	resp, err := self.userDataCenter.AddUser(newUser)
	if err != nil {
		self.writeErrorMessage(" add user get error: "+err.Error(), w)
		return
	}
	log.Println("process one user add req:" +
		r.Host + "! get resp: " + string(resp))
	fmt.Fprintf(w, "%s", resp)
}

func (self *UserServer) showAllUsersHandler(w http.ResponseWriter, r *http.Request) {
	resp, err := self.userDataCenter.UserList()
	if err != nil {
		self.writeErrorMessage("some err with user list: "+err.Error(), w)
		return
	}
	log.Println("process one show user list req:" +
		r.Host + "! get resp: " + string(resp))
	fmt.Fprintf(w, "%s", resp)
}

func (self *UserServer) writeErrorMessage(msg string, w http.ResponseWriter) {
	log.Println(msg)
	fmt.Fprintf(w, "%s", GetErrorMsg(msg))
}
