package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

var fs = http.FileServer(http.Dir("./public"))

// serves static files from a public directory. expects to be deployed behind a
// public-facing load balancer, like nginx, with TLS terminated prior at the
// load balancer.
func main() {
	lh := loggingHandler{}
	http.Handle("/", lh)

	fmt.Println("Listening on :1024...")
	err := http.ListenAndServe(":1024", lh)
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		os.Exit(1)
	}
}

type loggingHandler struct{}

func (lh loggingHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("[WEB] serving '%s'->'%s\n", req.URL.EscapedPath(), remoteAddr(req))
	fs.ServeHTTP(w, req)
}

func remoteAddr(req *http.Request) string {
	if addr := req.Header.Get("X-FORWARDED-FOR"); addr != "" {
		return addr
	}

	return req.RemoteAddr
}
