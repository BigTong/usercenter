package usercenter

import (
	"log"
	"net/http"
	"runtime"
	"testing"

	"github.com/gorilla/mux"
)

func init() {
	go MockServer()
}

func MockServer() {
	runtime.GOMAXPROCS(4)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	flag.Parse()

	service := NewUserServer()

	router := mux.NewRouter()
	router.HandleFunc("/users", service.UserRequestHandler)
	router.HandleFunc("/users/{userId:[0-9]+}/relationships",
		service.GetRelationshipHandler)
	router.HandleFunc("/users/{userId:[0-9]+}/relationships/{otherUserId:[0-9]+}",
		service.PutRelationshipHandler)
	log.Fatal(http.ListenAndServe("localhost:8080", router))
}

func BenchmarkGetUserList(b *testing.B) {
}

func BenchmarkAddUser(b *testing.B) {

}

func BenchmarkAddUserRelation(b *testing.B) {

}

func BenchmarkGetUserRelation(b *testing.B) {

}
