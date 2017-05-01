package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"usercenter"

	"github.com/BigTong/common/log"
	"github.com/gorilla/mux"
)

var (
	addr = flag.String("addr", "localhost:8080", "tcp addr")
)

func main() {
	runtime.GOMAXPROCS(4)
	logConfig := flag.String("log_file", "conf/log.json", "")
	flag.Parse()

	log.InitFileLoggerFromConfigFile(*logConfig, log.INFO)
	service := usercenter.NewUserServer()

	router := mux.NewRouter()
	router.HandleFunc("/users", service.UserRequestHandler)
	router.HandleFunc("/users/{userId:[0-9]+}/relationships",
		service.GetRelationshipHandler)
	router.HandleFunc("/users/{userId:[0-9]+}/relationships/{otherUserId:[0-9]+}",
		service.PutRelationshipHandler)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		select {
		case <-sigs:
			service.ShutdownRacefully()
			os.Exit(0)
		}
	}()

	log.FFatal("%s", http.ListenAndServe(*addr, router).Error())
}
