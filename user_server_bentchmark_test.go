package usercenter

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"

	"usercenter/db"
	"usercenter/user"

	"github.com/BigTong/common/http/client"
	"github.com/BigTong/common/log"
	"github.com/BigTong/common/rand"
	"github.com/gorilla/mux"
	"github.com/zheng-ji/goSnowFlake"
)

func MockServer() {
	runtime.GOMAXPROCS(1)
	logConfig := flag.String("log_file", "conf/log.json", "")
	flag.Parse()

	log.InitFileLoggerFromConfigFile(*logConfig, log.INFO)

	service := NewUserServer()

	router := mux.NewRouter()
	router.HandleFunc("/users", service.UserRequestHandler)
	router.HandleFunc("/users/{userId:[0-9]+}/relationships",
		service.GetRelationshipHandler)
	router.HandleFunc("/users/{userId:[0-9]+}/relationships/{otherUserId:[0-9]+}",
		service.PutRelationshipHandler)
	log.FFatal("%s", http.ListenAndServe("localhost:8080", router).Error())
}

var serverStarted = false
var mutex = &sync.Mutex{}

func StartServer() {
	mutex.Lock()
	defer mutex.Unlock()
	if serverStarted {
		return
	}
	go MockServer()
	time.Sleep(5 * time.Second)
	serverStarted = true
}

func BenchmarkGetUserList(b *testing.B) {
	StartServer()

	b.StopTimer()

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		client.GetDefaultHttpClient().Get(USER_API, nil, nil, nil, nil)

	}
}

func BenchmarkAddUserData(b *testing.B) {
	StartServer()
	b.StopTimer()

	b.StartTimer()

	httpClient := client.GetDefaultHttpClient()
	for i := 0; i < b.N; i++ {
		newUser := &user.User{
			Name: fmt.Sprintf("li_si_%d", time.Now().UnixNano()),
		}
		data, _ := json.Marshal(newUser)
		httpClient.JsonPost(USER_API, nil, data)

	}

}

func BenchmarkAddUserRelation(b *testing.B) {
	StartServer()
	b.StopTimer()

	b.StartTimer()
	httpClient := client.GetDefaultHttpClient()
	idWorker, _ := goSnowFlake.NewIdWorker(1)
	for i := 0; i < b.N; i++ {
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
		httpClient.JsonPut(url, nil, data)
	}
}

func LoadRelationsId() []int64 {
	postgresdb := db.NewPostgresQlDb(*postgresdbConfigFile)
	return postgresdb.GetAllUserRelationsId()
}

func BenchmarkGetUserRelation(b *testing.B) {
	StartServer()
	b.StopTimer()

	b.StartTimer()
	httpClient := client.GetDefaultHttpClient()
	idSet := LoadRelationsId()
	idSetLen := len(idSet)
	safeRand := rand.NewSafeRand()
	for i := 0; i < b.N; i++ {
		userId := fmt.Sprintf("%d",
			idSet[safeRand.Intn(idSetLen)])
		url := USER_API + "/" + userId + API_PATH_SEG
		httpClient.Get(url, nil, nil, nil, nil)

	}
}
