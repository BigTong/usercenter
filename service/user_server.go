package main

import (
	"flag"
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"

	"usercenter"

	"github.com/gorilla/mux"
)

var (
	addr = flag.String("addr", "localhost:8080", "tcp addr")
)

func main() {
	runtime.GOMAXPROCS(4)

	flag.Parse()

	service := usercenter.NewUserServer()

	router := mux.NewRouter()
	router.HandleFunc("/users", service.UserRequestHandler)
	router.HandleFunc("/users/{userId:[0-9]+}/relationships",
		service.GetRelationshipHandler)
	router.HandleFunc("/users/{userId:[0-9]+}/relationships/{otherUserId:[0-9]+}",
		service.PutRelationshipHandler)
	log.Fatal(http.ListenAndServe(*addr, router))
}
