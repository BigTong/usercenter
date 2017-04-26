package main

import (
	"flag"
	"log"

	"usercenter"

	"github.com/valyala/fasthttp"
)

var (
	addr = flag.String("addr", "localhost:8080", "tcp addr")
)

func main() {
	flag.Parse()

	userServer := usercenter.NewUserServer()
	requestHandler := func(ctx *fasthttp.RequestCtx) {
		userServer.RequestHandler(ctx)
	}

	if len(*addr) == 0 {
		log.Fatalf("not valid http server addr for %s", *addr)
	}

	fasthttp.ListenAndServe(*addr, requestHandler)
}
