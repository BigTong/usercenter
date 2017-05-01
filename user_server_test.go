package usercenter

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"usercenter/user"

	// "github.com/BigTong/common/file"
	"github.com/BigTong/common/http/client"
	// "github.com/BigTong/common/rand"
	// "github.com/gorilla/mux"
	"github.com/zheng-ji/goSnowFlake"
)

const (
	USER_API     = "http://localhost:8080/users"
	API_PATH_SEG = "/relationships"
)

func TestUserGetAddApi(t *testing.T) {
	StartServer()

	newUser := &user.User{
		Name: fmt.Sprintf("li_si_%d", time.Now().UnixNano()),
	}

	data, err := json.Marshal(newUser)
	if err != nil {
		t.Error(err.Error())
	}

	httpClient := client.GetDefaultHttpClient()
	resp, err := httpClient.JsonPost(USER_API, nil, data)

	respUser := &user.User{}
	err = json.Unmarshal(resp.Body, respUser)
	if err != nil {
		t.Error(err.Error())
	}

	if respUser.Name != newUser.Name {
		t.Error("name not same")
	}

	userList := []*user.User{}
	resp, err = httpClient.Get(USER_API, nil, nil, nil, nil)
	if err != nil {
		t.Error(err.Error())
	}

	err = json.Unmarshal(resp.Body, &userList)
	if err != nil {
		t.Error(err.Error())
	}

	if len(userList) == 0 {
		t.Error("get empty userlist")
	}
}

func TestUserRelationsApi(t *testing.T) {
	StartServer()
	idWorker, _ := goSnowFlake.NewIdWorker(1)
	iUserId, _ := idWorker.NextId()
	iOtherUserId, _ := idWorker.NextId()

	userId := fmt.Sprintf("%d", iUserId)
	otherUserId := fmt.Sprintf("%d", iOtherUserId)

	url := USER_API + "/" + userId +
		API_PATH_SEG + "/" + otherUserId
	relation := &user.UserRelationShip{
		State: user.RELATION_STATE_LIKED,
	}
	data, _ := json.Marshal(relation)

	httpClient := client.GetDefaultHttpClient()
	resp, err := httpClient.JsonPut(url, nil, data)
	if err != nil {
		t.Error(err.Error())
	}

	respRelation := &user.UserRelationShip{}
	err = json.Unmarshal(resp.Body, respRelation)
	if err != nil {
		t.Error(err.Error())
	}

	if fmt.Sprintf("%d", respRelation.Otherside) != otherUserId {
		t.Error("otherside id not equal")
	}

	// test matched yet
	url = USER_API + "/" + otherUserId +
		API_PATH_SEG + "/" + userId

	resp, err = httpClient.JsonPut(url, nil, data)
	if err != nil {
		t.Error(err.Error())
	}
	err = json.Unmarshal(resp.Body, respRelation)
	if err != nil {
		t.Error(err.Error())
	}
	if respRelation.State != user.RELATION_STATE_MATCHED {
		t.Error("state not matched")
	}

	// test get api
	url = USER_API + "/" + userId + API_PATH_SEG
	resp, err = httpClient.Get(url, nil, nil, nil, nil)
	if err != nil {
		t.Error(err.Error())
	}

	userRelatios := []*user.UserRelationShip{}
	err = json.Unmarshal(resp.Body, &userRelatios)
	if err != nil {
		t.Error(err.Error())
	}

	find := false
	for _, r := range userRelatios {
		if fmt.Sprintf("%d", r.Otherside) == otherUserId {
			find = true
			break
		}
	}

	if !find {
		t.Error("not find add user relation")
	}
}
