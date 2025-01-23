package main

import (
	"fmt"
	"net/http"
	"os"
)

// serves static files from a public directory. expects to be deployed behind a
// public-facing load balancer, like nginx, with TLS terminated prior at the
// load balancer.
func main() {
	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/", fs)

	fmt.Println("Listening on :1024...")
	err := http.ListenAndServe(":1024", fs)
	if err != nil {
		fmt.Printf("failed to listen: %v\n", err)
		os.Exit(1)
	}
}
