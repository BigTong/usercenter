package usercenter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"usercenter/user"

	"github.com/valyala/fasthttp"
	"github.com/zheng-ji/goSnowFlake"
)

const (
	SERVER_PATH = "/users"
)

type ErrResp struct {
	Error string `json:"error"`
}

func GetErrorMsg(errMsg string) string {
	errResp := &ErrResp{errMsg}
	data, _ := json.Marshal(errResp)
	return string(data)
}

func NewUserServer() *UserServer {
	idGenerater, err := goSnowFlake.NewIdWorker(1)
	if err != nil {
		log.Fatalf("new id generater get error:%s", err.Error())
	}

	return &UserServer{
		userDataCenter: NewUserDataCenter(),
		idGenerater:    idGenerater,
	}
}

type UserServer struct {
	userDataCenter *UserDataCenter
	idGenerater    *goSnowFlake.IdWorker
}

func (self *UserServer) RequestHandler(ctx *fasthttp.RequestCtx) {
	if string(ctx.Path()) != SERVER_PATH {
		fmt.Fprintf(ctx.Response.BodyWriter(),
			"not valid path for: %s",
			GetErrorMsg(string(ctx.Path())))
		return
	}

	if ctx.IsGet() {
		self.userListHandler(ctx)
		return
	}

	if ctx.IsPost() {
		self.userAddHandler(ctx)
		return
	}
	errMsg := "not support method: " + string(ctx.Method())
	log.Println(errMsg)
	fmt.Fprintf(ctx.Response.BodyWriter(), "%s", errMsg)
}

type UserPostData struct {
	Name string `json:"name"`
}

func (self *UserServer) userAddHandler(ctx *fasthttp.RequestCtx) {
	body := ctx.Request.Body()
	decoder := json.NewDecoder(bytes.NewReader(body))
	userPostData := &UserPostData{}
	if err := decoder.Decode(userPostData); err != nil {
		errMsg := "decode post data get error:" + err.Error()
		log.Println(errMsg)
		log.Println("body: " + string(body))
		fmt.Fprintf(ctx.Response.BodyWriter(), "%s", GetErrorMsg(errMsg))
		return
	}

	userName := userPostData.Name
	ok := self.userDataCenter.CheckValidAndUpdateForUser(userName)
	if !ok {
		errMsg := "not valid user name: " + userName + ", please change user name!"
		log.Println(errMsg)
		fmt.Fprintf(ctx.Response.BodyWriter(), "%s", GetErrorMsg(errMsg))
		return
	}

	id, err := self.idGenerater.NextId()
	if err != nil {
		errMsg := "id generater get error: " + err.Error()
		log.Println(errMsg)
		fmt.Fprintf(ctx.Response.BodyWriter(), "%s", errMsg)
		return
	}

	newUser := &user.User{
		Name:        userName,
		Id:          id,
		CreatedTime: time.Now().Unix(),
	}
	resp, err := self.userDataCenter.AddUser(newUser)
	if err != nil {
		errMsg := " add user get error: " + err.Error()
		log.Println(errMsg)
		fmt.Fprintf(ctx.Response.BodyWriter(), "%s", errMsg)
		return
	}
	log.Println("process one user add req:" +
		string(ctx.Host()) + "! get resp: " + string(resp))
	fmt.Fprintf(ctx.Response.BodyWriter(), "%s", resp)
}

func (self *UserServer) userListHandler(ctx *fasthttp.RequestCtx) {
	resp, err := self.userDataCenter.UserList()
	if err != nil {
		errMsg := "some err with user list: " + err.Error()
		log.Println(errMsg)
		fmt.Fprintf(ctx.Response.BodyWriter(), "%s", GetErrorMsg(errMsg))
		return
	}
	log.Println("process one show user list req:" +
		string(ctx.Host()) + "! get resp: " + string(resp))
	fmt.Fprintf(ctx.Response.BodyWriter(), "%s", resp)
}
